package main

import (
	"fmt"
	"os"
	"time"

	"github.com/indexooor/core/indexooor"
	"github.com/urfave/cli/v2"
)

var (
	indexCommand = &cli.Command{
		Name:   "index",
		Usage:  "Index a contract",
		Action: startIndexing,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "address", Usage: "Contract address"},
			&cli.Int64Flag{Name: "start-block", Usage: "Block to start indexing from"},
			// TODO: add more flags as needed
		},
	}

	// TODO: add more commands as needed

	rpcFlag = &cli.StringFlag{
		Name:  "rpc",
		Usage: "RPC endpoint of the node",
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

	fmt.Println("Starting indexing...")

	// Start a new go routine to index
	go indexooor.StartIndexing()

	time.Sleep(time.Second * 5)

	return nil
}
