package indexooor

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// StartIndexing starts indexing a contract address
func StartIndexing(_rpc string, startBlock uint64, contractAddresses []string) error {

	// Expected inputs: contract address, rpc

	// This logic might work for single contract as of now.
	// Build a generic logic which should work for multiple contracts in single iteration

	// TODOs
	/*
	 * 1. Find contract creation block (or maybe ask for it from user?)
	 * 2. Start querying every block and call eth_getStorageRoot for that contract
	 * 3. If contract root is changed, fetch txs, iterate on them and get debug_traceTransaction state diff
	 * 4. Interact with DB to push values if required
	 */

	// initialise data necessary for indexing

	contractStorageHashes := make(map[string]string)

	currentBlock := startBlock

	for i := 0; i < len(contractAddresses); i++ {
		contractStorageHashes[contractAddresses[i]] = ""
	}

	rpc, _ := rpc.Dial(_rpc)
	gethclient := gethclient.New(rpc)
	client := ethclient.NewClient(rpc)

	// forever loop
	for {
		// get current block number
		latestBlockNumber := getBlockNumber(client)

		indexBlock := false

		// make a list of contracts to index
		contractsToIndex := []string{}

		// if current block number is not equal to latest block, index data
		if currentBlock != latestBlockNumber {
			// iterate over all contracts and call getProof and get storage hash
			for i := 0; i < len(contractAddresses); i++ {
				// get storage hash
				storageHash := getProof(gethclient, contractAddresses[i], nil).StorageHash.Hex()

				// if not equal to previous storage hash, index data
				if storageHash != contractStorageHashes[contractAddresses[i]] {
					indexBlock = true
					contractsToIndex = append(contractsToIndex, contractAddresses[i])
				}
			}

			// if indexBlock is true, index data
			if indexBlock {
				// get block by number
				block := getBlockByNumber(client, nil)

				// iterate over all transactions in the block
				for i := 0; i < block.Transactions().Len(); i++ {
					// get transaction hash
					txHash := block.Transactions()[i].Hash().Hex()

					// call debug_traceTransaction

					txnTrace := debugTraceTransaction(rpc, txHash)

					// iterate over contracts and check trace, if trace has post for contract address, store to db
					for j := 0; j < len(contractsToIndex); j++ {
						if txnTrace["post"][contractsToIndex[j]] != nil {
							// store to db
							v := txnTrace["post"][contractsToIndex[j]]
							storage := v.(map[string]interface{})["storage"]
							fmt.Println("Post storage for", contractsToIndex[j], storage)

							// iterate over all keys in storage and store to db with slot id as key
							for k, v := range storage.(map[string]interface{}) {
								fmt.Println("Key", k, "Value", v)
							}

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

// indexing function where all contracts are indexed in full mode
func StartIndexingFullMode(_rpc string, startBlock uint64) {

	currentBlock := startBlock

	rpc, _ := rpc.Dial(_rpc)

	client := ethclient.NewClient(rpc)

	// forever loop
	for {
		// get current block number
		latestBlockNumber := getBlockNumber(client)
		// if current block number is not equal to latest block, index data
		if currentBlock != latestBlockNumber {

			// get block by number
			block := getBlockByNumber(client, nil)

			// iterate over all transactions in the block
			for i := 0; i < block.Transactions().Len(); i++ {
				// get transaction hash
				txHash := block.Transactions()[i].Hash().Hex()

				// call debug_traceTransaction

				txnTrace := debugTraceTransaction(rpc, txHash)

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

// AccountResult is the result of a GetProof operation.
func getProof(client *gethclient.Client, contractAddress string, blockNumber *big.Int) *gethclient.AccountResult {

	result, err := client.GetProof(context.Background(), common.HexToAddress(contractAddress), []string{}, blockNumber)

	// finalizedBlock, err := client.BlockByNumber(context.Background(), new(big.Int).SetInt64(29652208)) // 0x190B6C0

	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("Storage Hash", result.StorageHash)
	fmt.Println("Account Address", result.Address)

	return result
}

func getBlockByNumber(client *ethclient.Client, blockNumber *big.Int) *types.Block {

	results, err := client.BlockByNumber(context.Background(), blockNumber)

	if err != nil {
		fmt.Println("err", err)
	}

	// fmt.Println("Len Of Transactions", results.Transactions().Len())

	// for i := 0; i < results.Transactions().Len(); i++ {
	// 	fmt.Println("Transaction", results.Transactions()[i].Hash().Hex())
	// }

	return results

}

func getBlockNumber(client *ethclient.Client) uint64 {
	results, err := client.BlockNumber(context.Background())
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("results", results)

	return results
}

func debugTraceTransaction(rpcClient *rpc.Client, txHash string) map[string]map[string]interface{} {
	// map[post: map[address: Account], pre: map[address: Account]]
	var result map[string]map[string]interface{}

	err := rpcClient.CallContext(context.Background(), &result, "debug_traceTransaction", txHash, map[string]interface{}{
		"tracer": "prestateTracer",
		"tracerConfig": map[string]interface{}{
			"diffMode": true,
		},
	})
	if err != nil {
		fmt.Println("err", err)
	}

	return result

	// for k, v := range result {
	// 	if k == "post" {
	// 		addr := "0x3126d03e98bb95a7d4046ba8a64369e6656fe448"
	// 		storage := v[addr].(map[string]interface{})["storage"]
	// 		fmt.Println("Post storage for", addr, storage)
	// 	}
	// }
}

// func main9() {
// 	fmt.Println("Hello, playground")
// 	rpc, _ := rpc.Dial("https://eth-goerli-rpc.gateway.pokt.network/")
// 	gethclient := gethclient.New(rpc)
// 	client := ethclient.NewClient(rpc)

// 	getProof(gethclient, "0x17fCb0e5562c9f7dBe2799B254e0948568973B36", nil)
// 	getBlockByNumber(client, nil)
// 	getBlockNumber(client)

// 	debugTraceTransaction(rpc, "0xfd70fc72a37a912426957581f8923c0f7f24d938c8bcbeed45f82d083f8ad745")

// }
