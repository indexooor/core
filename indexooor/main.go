package indexooor

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// StartIndexing starts indexing a contract address
func StartIndexing() error {
	// Expected inputs: contract address, rpc

	fmt.Println("Starting indexing... 2")

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

func main9() {
	fmt.Println("Hello, playground")
	rpc, _ := rpc.Dial("https://eth-goerli-rpc.gateway.pokt.network/")
	client := gethclient.New(rpc)
	getProof(client, "0x17fCb0e5562c9f7dBe2799B254e0948568973B36", nil)
}
