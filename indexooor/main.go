package indexooor

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"

	log "github.com/inconshreveable/log15"

	database "github.com/indexooor/core/db"
)

var errInvalidStartBlock error = errors.New("invalid start block")

// StartIndexing is the main loop which starts indexing the given contract addresses
// from a start block using an RPC endpoint.
func StartIndexing(_rpc string, startBlock uint64, contractAddresses []string, runId uint64) error {
	// Setup the DB
	db, err := database.SetupDB()
	if err != nil {
		return err
	}
	defer db.Close()

	var run *database.Run

	if runId != 0 {
		// Look out for an existing run by ID
		log.Info("Looking for an existing run entry for indexing data", "id", runId)
		run, err = db.FetchRunByID(runId)
	}

	if run == nil || err != nil {
		log.Info("Creating a new run entry for indexing data")
		// Create a new run in table
		run = &database.Run{
			StartBlock: startBlock,
			LastBlock:  startBlock,
			Contracts:  contractAddresses,
		}
		err = db.CreateNewRun(run)
		if err != nil {
			return err
		}
	}

	// TODO: Fetch latest run ID
	log.Info("Using run", "id", run.Id)

	// Use contracts from runs table
	contractAddresses = run.Contracts

	// initialise data necessary for indexing
	var (
		currentBlock          uint64 = startBlock
		latestBlockNumber     uint64
		contractStorageHashes = make(map[string]common.Hash)
	)

	for i := 0; i < len(contractAddresses); i++ {
		// Initialise with empty storage root
		contractStorageHashes[contractAddresses[i]] = common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
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

				log.Debug("Fetched storage root", "number", currentBlock, "root", storageRoot, "existing", contractStorageHashes[contractAddresses[i]])

				// if not equal to previous storage hash, index data
				if storageRoot != contractStorageHashes[contractAddresses[i]] {
					indexBlock = true
					contractsToIndex = append(contractsToIndex, contractAddresses[i])
					contractStorageHashes[contractAddresses[i]] = storageRoot // Update storage root
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
						contractTrace := txnTrace["post"][strings.ToLower(contractsToIndex[j])]
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
							log.Debug("Inserting data into DB", "contract", obj.Contract, "slot", obj.Slot, "value", obj.Value)
							db.AddNewIndexingEntry(obj)
							log.Debug("Done adding data")
						}
					}
				}
			} else {
				log.Info("Nothing to index", "block", currentBlock)
			}

			// Update the last block indexed in the run (can be done async)
			// and ignore err for now
			db.UpdateRun(run.Id, currentBlock)

			currentBlock++
		} else {
			log.Info("Indexed till tip of chain, waiting for 10s", "block", currentBlock)
			time.Sleep(10 * time.Second)
		}
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
