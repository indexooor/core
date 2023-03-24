package indexooor

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	database "github.com/indexooor/core/db"
)

var errInvalidStartBlock error = errors.New("invalid start block")

// StartIndexing is the main loop which starts indexing the given contract addresses
// from a start block using an RPC endpoint.
func StartIndexing(_rpc string, startBlock uint64, contractAddresses []string) error {
	// Setup the DB
	db, err := database.SetupDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// TODO: If a run ID is provided, use it else create a new one.

	// Create a new run in table
	run := &database.Run{
		StartBlock: int(startBlock),
		EndBlock:   int(startBlock),
		Contracts:  contractAddresses,
	}
	err = db.CreateNewRun(run)
	if err != nil {
		return err
	}

	// initialise data necessary for indexing

	var (
		currentBlock          uint64 = startBlock
		latestBlockNumber     uint64
		contractStorageHashes = make(map[string]common.Hash)
	)

	for i := 0; i < len(contractAddresses); i++ {
		contractStorageHashes[contractAddresses[i]] = common.Hash{}
	}

	rpc, _ := rpc.Dial(_rpc)
	gethclient := gethclient.New(rpc)
	client := ethclient.NewClient(rpc)

	latestBlockNumber, err = getBlockNumber(client)
	if err != nil {
		return err
	}

	// Return if start block is ahead of latest chain head
	if currentBlock > latestBlockNumber {
		log.Error("Latest chain head is behind provided start block, stopping", "chain head", latestBlockNumber, "start block", currentBlock)
		return errInvalidStartBlock
	}

	// forever loop
	for {
		// Update the upper limit if required
		if currentBlock > latestBlockNumber {
			// get current block number
			latestBlockNumber, err = getBlockNumber(client)
			if err != nil {
				return err
			}
		}

		indexBlock := false

		// make a list of contracts to index
		contractsToIndex := make([]string, 0, len(contractAddresses))

		// If we're behind latest block, index data
		if currentBlock <= latestBlockNumber {
			// iterate over all contracts and call getProof and get storage hash
			log.Info("Checking for storage root change in contracts", "block", currentBlock)
			for i := 0; i < len(contractAddresses); i++ {
				// get storage root
				storageRoot, err := getStorageRoot(gethclient, contractAddresses[i], big.NewInt(int64(currentBlock)))
				if err != nil {
					return err
				}

				// if not equal to previous storage hash, index data
				if storageRoot != contractStorageHashes[contractAddresses[i]] {
					indexBlock = true
					contractsToIndex = append(contractsToIndex, contractAddresses[i])
				}
			}

			// if indexBlock is true, index data
			if indexBlock {
				log.Info("State diff found against block, fetching traces for each block tx", "block", currentBlock)

				// get block by number
				block, err := getBlockByNumber(client, big.NewInt(int64(currentBlock)))
				if err != nil {
					return err
				}

				txs := block.Transactions()

				// iterate over all transactions in the block
				for i := 0; i < txs.Len(); i++ {
					// get transaction hash
					txHash := txs[i].Hash().Hex()

					// call debug_traceTransaction
					txnTrace, err := debugTraceTransaction(rpc, txHash)
					if err != nil {
						return err
					}

					// iterate over contracts and check trace, if trace has post for contract address, store to db
					for j := 0; j < len(contractsToIndex); j++ {
						contractTrace := txnTrace["post"][contractsToIndex[j]]
						if contractTrace == nil {
							continue
						}

						// access the storage field of the contract
						storage := contractTrace.(map[string]interface{})["storage"]

						if storage == nil {
							continue
						}

						// iterate over all keys in storage and store to db with (slot + contract address) -> value
						for slot, value := range storage.(map[string]interface{}) {
							obj := &database.Indexooor{
								Slot:     slot,
								Value:    value.(string),
								Contract: contractsToIndex[j],
							}
							log.Info("Inserting data into DB", "contract", obj.Contract, "slot", obj.Slot, "value", obj.Value)
							db.AddNewIndexingEntry(obj)
							log.Info("Done adding data")
						}
					}
				}
			} else {
				log.Info("Nothing to index", "block", currentBlock)
			}
		} else {
			log.Info("Indexed till tip of chain, waiting for 10s")
			time.Sleep(10 * time.Second)
		}

		currentBlock++
	}
}

// StartIndexingFullMode indexes all the existing contracts on a network from a start block.
// TODO: Refactor according to startIndexing else won't work as of now.
func StartIndexingFullMode(_rpc string, startBlock uint64) {
	currentBlock := startBlock

	rpc, _ := rpc.Dial(_rpc)

	client := ethclient.NewClient(rpc)

	// forever loop
	for {
		// get current block number
		latestBlockNumber, _ := getBlockNumber(client)
		// if current block number is not equal to latest block, index data
		if currentBlock != latestBlockNumber {

			// get block by number
			block, _ := getBlockByNumber(client, nil)

			// iterate over all transactions in the block
			for i := 0; i < block.Transactions().Len(); i++ {
				// get transaction hash
				txHash := block.Transactions()[i].Hash().Hex()

				// call debug_traceTransaction

				txnTrace, _ := debugTraceTransaction(rpc, txHash)

				// iterate over all keys in trace and check trace, if trace has post for contract address, store to db
				for k := range txnTrace["post"] {

					if txnTrace["post"][k] != nil {
						// store to db
						v := txnTrace["post"][k]
						storage := v.(map[string]interface{})["storage"]
						fmt.Println("Post storage for", k, storage)

						// iterate over all keys in storage and store to db with slot id as key
						for k, v := range storage.(map[string]interface{}) {
							fmt.Println("Key", k, "Value", v)
						}

					}
				}
			}

		} else {

			// sleep for 8 seconds
			time.Sleep(time.Second * 8)

		}

	}
}

// getStorageRoot calls the eth_getProof endpoint against a contract address and block number
// and returns the storage trie root for the contract.
func getStorageRoot(client *gethclient.Client, address string, number *big.Int) (common.Hash, error) {
	proof, err := client.GetProof(context.Background(), common.HexToAddress(address), []string{}, number)
	if err != nil {
		log.Info("Error fetching proof of contract address", "address", address, "number", number, "err", err)
		return common.Hash{}, err
	}

	return proof.StorageHash, nil
}

// getBlockByNumber returns block by number
func getBlockByNumber(client *ethclient.Client, number *big.Int) (*types.Block, error) {
	block, err := client.BlockByNumber(context.Background(), number)
	if err != nil {
		log.Error("Error fetching block by number", "number", number, "err", err)
	}

	return block, nil
}

// getBlockNumber returns latest block number
func getBlockNumber(client *ethclient.Client) (uint64, error) {
	number, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Error("Error fetching chain block number", "err", err)
		return 0, err
	}

	return number, err
}

// debugTraceTransaction calls debug_traceTransaction for the given transaction hash and uses "preStateTracer" to fetch
// state diff.
func debugTraceTransaction(rpcClient *rpc.Client, txHash string) (map[string]map[string]interface{}, error) {
	// map[post: map[address: Account], pre: map[address: Account]]
	var result map[string]map[string]interface{} // TODO: Convert this into a struct based object

	err := rpcClient.CallContext(context.Background(), &result, "debug_traceTransaction", txHash, map[string]interface{}{
		"tracer": "prestateTracer",
		"tracerConfig": map[string]interface{}{
			"diffMode": true,
		},
	})

	if err != nil {
		log.Error("Error in running debug trace transaction", "hash", txHash, "err", err)
	}

	// for k, v := range result {
	// 	if k == "post" {
	// 		addr := "0x3126d03e98bb95a7d4046ba8a64369e6656fe448"
	// 		storage := v[addr].(map[string]interface{})["storage"]
	// 		fmt.Println("Post storage for", addr, storage)
	// 	}
	// }

	return result, err
}
