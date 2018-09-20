package wallet

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// CLI responsible for processing command line arguments
type WalletCli struct{}

func (wallet *WalletCli) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
}

func (wallet *WalletCli) validateArgs() {
	if len(os.Args) < 2 {
		wallet.printUsage()
		os.Exit(1)
	}
}

// Run parses command line arguments and processes commands
func (wallet *WalletCli) Run() {
	wallet.validateArgs()
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	nodeID := os.Args[3]
	if nodeID == "" {
		fmt.Printf("NODE_ID not set!")
		return
	}
	if 5 > len(os.Args){
		fmt.Printf("Usage not set!")
		return
	}

	switch os.Args[4] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[4:])
		if err != nil {
			log.Panic(err)
		}

	case "createwallet":
		err := createWalletCmd.Parse(os.Args[4:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[4:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[4:])
		if err != nil {
			log.Panic(err)
		}
	default:
		wallet.printUsage()
		os.Exit(1)
	}

	if createWalletCmd.Parsed() {
		wallet.createWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		wallet.listAddresses(nodeID)
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		wallet.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}
}