package main

import (
	"os"
	"strings"

	"github.com/indexooor/core/indexooor"
	"github.com/urfave/cli/v2"

	log "github.com/inconshreveable/log15"
)

var (
	indexCommand = &cli.Command{
		Name:   "index",
		Usage:  "Index a contract",
		Action: startIndexing,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "rpc",
				Usage:       "RPC endpoint of the node",
				Value:       "http://localhost:8545",
				DefaultText: "http://localhost:8545",
				Required:    true,
			},
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
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "Enable debug logs",
				Value:       false,
				DefaultText: "false",
			},
		},
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

	if err := app.Run(os.Args); err != nil {
		log.Error("Error in indexer service, exiting", "err", err)
		os.Exit(1)
	}
}

func startIndexing(ctx *cli.Context) error {
	startBlock := ctx.Uint64("start-block")
	rpc := ctx.String("rpc")
	runId := ctx.Uint64("run-id")

	if ctx.Bool("debug") {
		log.Root().SetHandler(log.LvlFilterHandler(4, log.StreamHandler(os.Stderr, log.TerminalFormat())))
	}

	// split string by comma
	contractAddresses := strings.Split(ctx.String("contract-addresses"), ",")
	for i, contractAddress := range contractAddresses {
		contractAddresses[i] = strings.TrimSpace(contractAddress)
	}

	log.Info("Indexoooor goes vrooom vrooom 🚀🚀")

	log.Info("Starting to index", "contracts", contractAddresses, "start block", startBlock, "rpc", rpc, "run-id", runId)

	return indexooor.StartIndexing(rpc, startBlock, contractAddresses, runId)
}
