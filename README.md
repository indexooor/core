# Indexooor core
Core indexooor that indexes your contracts data into database for fast data availability

# Usage
Building the binary
```
make build
```

This will create an `indexooor` binary in your `$GOBIN`. Use the binary with flags for indexing. 
```
indexooor index <flags>
```

Refer to the help command `--help` for flag details.
```
indexooor index --help

NAME:
   Indexooor index - Index a contract

USAGE:
   Indexooor index [command options] [arguments...]

OPTIONS:
   --rpc value                 RPC endpoint of the node (default: http://localhost:8545)
   --contract-addresses value  Contract address (comma seperated for multiple contracts) (default: "")
   --start-block value         Block to start indexing from (default: 0)
   --run-id value              Run ID to start indexing from block where left off (default: 0)
   --debug                     Enable debug logs (default: false)
   --help, -h                  show help
```
