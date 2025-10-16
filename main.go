// In main.go

package main

import (
	"github.com/Triad-0112/BlockChain.git/blockchain" // Change to your module path
	"github.com/Triad-0112/BlockChain.git/cli"        // Change to your module path
)

func main() {
	bc := blockchain.NewBlockchain()
	defer bc.CloseDB() // This line is new! It ensures the DB is closed.

	cli := cli.NewCLI(bc)
	cli.Run()
}
