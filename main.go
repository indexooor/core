package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/indexooor/core/indexooor"
	"github.com/urfave/cli/v2"
)

var (
	indexCommand = &cli.Command{
		Name:   "index",
		Usage:  "Index a contract",
		Action: startIndexing,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "contract-addresses",
				Usage:       "Contract address (comma seperated for multiple contracts)",
				Value:       "",
				DefaultText: "0x17fCb0e5562c9f7dBe2799B254e0948568973B36,0x17fCb0e5562c9f7dBe2799B254e0948568973B34",
				Required:    true,
			},
			&cli.Int64Flag{
				Name:        "start-block",
				Usage:       "Block to start indexing from",
				Value:       0,
				DefaultText: "0",
			},
			// TODO: add more flags as needed
		},
	}

	// TODO: add more commands as needed

	rpcFlag = &cli.StringFlag{
		Name:        "rpc",
		Usage:       "RPC endpoint of the node",
		Value:       "https://eth-goerli-rpc.gateway.pokt.network/",
		DefaultText: "https://eth-goerli-rpc.gateway.pokt.network/",
	}

	// (TODO): Add more flags as needed
)

func main() {
	// Visit https://cli.urfave.org/v2/examples/flags/ for references
	app := cli.NewApp()
	app.Name = "OP Indexooor"
	app.Usage = "A command-line utility to interact with the indexer service"
	app.Commands = []*cli.Command{
		indexCommand,
	}
	app.Flags = []cli.Flag{
		rpcFlag,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startIndexing(ctx *cli.Context) error {
	// (TODO): Perform checks on flag values here

	fmt.Println("Indexing For")

	// Print all flags
	fmt.Println("Contract addresses:", ctx.String("contract-addresses"))
	fmt.Println("Start block:", ctx.Int64("start-block"))
	fmt.Println("RPC:", ctx.String("rpc"))

	startBlock := ctx.Int64("start-block")
	rpc := ctx.String("rpc")

	// split string by comma
	contractAddresses := strings.Split(ctx.String("contract-addresses"), ",")
	// split all strings of spaces
	for i, contractAddress := range contractAddresses {
		contractAddresses[i] = strings.TrimSpace(contractAddress)
	}

	fmt.Println("Starting indexing...")

	// Start a new go routine to index
	indexooor.StartIndexing(rpc, startBlock, contractAddresses)

	return nil
}
