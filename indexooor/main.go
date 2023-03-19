package indexooor

// StartIndexing starts indexing a contract address
func StartIndexing() error {
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

	return nil
}

// import (
// 	"context"
// 	"fmt"
// 	"math/big"

// 	"github.com/maticnetwork/bor/ethclient"
// 	"github.com/maticnetwork/bor/rpc"
// )

// func getBlock(client *ethclient.Client) {
// 	finalizedBlock, err := client.BlockByNumber(context.Background(), new(big.Int).SetInt64(29652208)) // 0x190B6C0
// 	if err != nil {
// 		fmt.Println("err", err)
// 	}
// 	fmt.Println("number", finalizedBlock.NumberU64(), "finalizedBlock.Header.TxHash", finalizedBlock.Header().TxHash)
// }

// func main9() {
// 	rpc, _ := rpc.Dial("")
// 	client := ethclient.NewClient(rpc)
// 	getBlock(client)
// }
