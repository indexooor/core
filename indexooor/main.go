package indexooor

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// StartIndexing starts indexing a contract address
func StartIndexing(rpc string, startBlock int64, contractAddresses []string) error {
	// Expected inputs: contract address, rpc

	// This logic might work for single contract as of now.
	// Build a generic logic which should work for multiple contracts in single iteration

	main9()

	// TODOs
	/*
	 * 1. Find contract creation block (or maybe ask for it from user?)
	 * 2. Start querying every block and call eth_getStorageRoot for that contract
	 * 3. If contract root is changed, fetch txs, iterate on them and get debug_traceTransaction state diff
	 * 4. Interact with DB to push values if required
	 */

	return nil
}

func getProof(client *gethclient.Client, contractAddress string, blockNumber *big.Int) {

	result, err := client.GetProof(context.Background(), common.HexToAddress(contractAddress), []string{}, blockNumber)

	// finalizedBlock, err := client.BlockByNumber(context.Background(), new(big.Int).SetInt64(29652208)) // 0x190B6C0

	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("Storage Hash", result.StorageHash)
	fmt.Println("Account Address", result.Address)
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

type Account struct {
	balance string            `json:"balance"`
	code    string            `json:"code"`
	nonce   uint64            `json:"nonce"`
	storage map[string]string `json:"storage"`
}
type DebugTraceTransactionResult struct {
	pre  map[string]map[string]Account `json:"pre"`
	post map[string]map[string]Account `json:"post"`
}

func debugTraceTransaction(rpcClient *rpc.Client, txHash string) DebugTraceTransactionResult {

	var result DebugTraceTransactionResult
	err := rpcClient.CallContext(context.Background(), &result, "debug_traceTransaction", txHash, map[string]interface{}{
		"tracer": "prestateTracer",
		"tracerConfig": map[string]interface{}{
			"diffMode":    true,
			"onlyTopCall": true,
		},
	})
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("result", result)

	fmt.Printf("pre: %+v\n", result)
	// fmt.Printf("post: %#v\n", result.post);

	return result
}

func main9() {
	fmt.Println("Hello, playground")
	rpc, _ := rpc.Dial("https://eth-goerli-rpc.gateway.pokt.network/")
	gethclient := gethclient.New(rpc)
	client := ethclient.NewClient(rpc)

	getProof(gethclient, "0x17fCb0e5562c9f7dBe2799B254e0948568973B36", nil)
	getBlockByNumber(client, nil)
	getBlockNumber(client)

	debugTraceTransaction(rpc, "0xfd70fc72a37a912426957581f8923c0f7f24d938c8bcbeed45f82d083f8ad745")

}
