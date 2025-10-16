// In cli/cli.go

package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Triad-0112/BlockChain.git/blockchain" // Change to your module path
)

// CLI is responsible for processing command line arguments.
type CLI struct {
	bc *blockchain.Blockchain
}

// NewCLI creates a new CLI instance.
func NewCLI(bc *blockchain.Blockchain) *CLI {
	return &CLI{bc}
}

// printUsage displays how to use the CLI.
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -data BLOCK_DATA    - add a block to the blockchain")
	fmt.Println("  printchain                   - print all the blocks of the blockchain")
}

// validateArgs ensures that the CLI was given valid arguments.
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run parses command line arguments and executes the appropriate command.
func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

// addBlock is the handler for the 'addblock' command.
func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
	fmt.Println("Success! A new block has been added.")
}

// printChain is the handler for the 'printchain' command.
func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()
		if block == nil {
			break
		}

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break // Reached the genesis block
		}
	}
}
