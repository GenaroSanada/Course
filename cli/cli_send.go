package main

import (
	"fmt"
	"log"
	"Course/cli/wallet"
)

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !Wallet.ValidateAddress(from) {
		fmt.Println("ERROR: Sender address is not valid")
		return
	}
	if !Wallet.ValidateAddress(to) {
		fmt.Println("ERROR: Recipient address is not valid")
		return
	}

	wallets, err := Wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	if string(wallet.GetAddress()[:]) != from {
		fmt.Println("error!")
	}
	//@todo Transaction
	fmt.Println(string(wallet.GetAddress()[:]))
	fmt.Println("Success!")
}
