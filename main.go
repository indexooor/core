package main

import (
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/log"
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
			&cli.Int64Flag{
				Name:        "run-id",
				Usage:       "Run ID to start indexing from block where left off",
				Value:       0,
				DefaultText: "0",
			},
		},
	}

	// Generic flag
	rpcFlag = &cli.StringFlag{
		Name:        "rpc",
		Usage:       "RPC endpoint of the node",
		Value:       "https://eth-goerli-rpc.gateway.pokt.network/",
		DefaultText: "https://eth-goerli-rpc.gateway.pokt.network/",
	}
)

func main() {
	// Visit https://cli.urfave.org/v2/examples/flags/ for references
	app := cli.NewApp()
	app.Name = "Indexooor"
	app.Usage = "A command-line utility to interact with the indexer service"
	app.Commands = []*cli.Command{
		indexCommand,
	}
	app.Flags = []cli.Flag{
		rpcFlag,
	}

	if err := app.Run(os.Args); err != nil {
		log.Error("Error in indexer service, exiting", "err", err)
		os.Exit(1)
	}
}

func startIndexing(ctx *cli.Context) error {
	startBlock := ctx.Uint64("start-block")
	rpc := ctx.String("rpc")
	runId := ctx.Uint64("run-id")

	// split string by comma
	contractAddresses := strings.Split(ctx.String("contract-addresses"), ",")
	for i, contractAddress := range contractAddresses {
		contractAddresses[i] = strings.TrimSpace(contractAddress)
	}

	log.Info("Indexoooor goes vrooom vrooom 🚀🚀")

	log.Info("Starting to index", "contracts", contractAddresses, "start block", startBlock, "rpc", rpc, "run-id", runId)

	return indexooor.StartIndexing(rpc, startBlock, contractAddresses, runId)
}
