package cli

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Triad-0112/BlockChain.git/blockchain"
	"github.com/Triad-0112/BlockChain.git/wallet"
)

type CLI struct{}

func NewCLI() *CLI {
	return &CLI{}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet      - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  listaddresses     - Lists all addresses from the wallet file")
	fmt.Println("  printchain        - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}
	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
}

func (cli *CLI) createBlockchain(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.NewBlockchain(address)
	defer bc.CloseDB()
	fmt.Println("Done! Blockchain created.")
}

func (cli *CLI) createWallet() {
	wallets, _ := wallet.NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) listAddresses() {
	wallets, err := wallet.NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) printChain() {
	bc := blockchain.NewBlockchain("")
	defer bc.CloseDB()
	bci := bc.Iterator()
	for {
		block := bci.Next()
		if block == nil {
			break
		}
		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %t\n", pow.Validate())
		fmt.Println("Transactions:")
		for _, tx := range block.Transactions {
			fmt.Printf("  - Transaction %x\n", tx.ID)
		}
		fmt.Printf("\n")
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) getBalance(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.NewBlockchain(address)
	defer bc.CloseDB()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	if !wallet.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchain(from)
	defer bc.CloseDB()

	tx, err := blockchain.NewUTXOTransaction(from, to, amount, bc)
	if err != nil {
		log.Panic(err)
	}
	bc.MineBlock([]*blockchain.Transaction{tx})
	fmt.Println("Success! Transaction sent.")
}
