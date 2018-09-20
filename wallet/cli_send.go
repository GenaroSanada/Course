package wallet

import (
	"fmt"
	"log"
)

func (wallet *WalletCli) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		fmt.Println("ERROR: Sender address is not valid")
		return
	}
	if !ValidateAddress(to) {
		fmt.Println("ERROR: Recipient address is not valid")
		return
	}

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet12 := wallets.GetWallet(from)
	if string(wallet12.GetAddress()[:]) != from {
		fmt.Println("error!")
	}
	//@todo Transaction
	fmt.Println(string(wallet12.GetAddress()[:]))
	fmt.Println("Success!")
}
