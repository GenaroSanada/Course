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
	wall := wallets.GetWallet(from)
	if string(wall.GetAddress()[:]) != from {
		fmt.Println("error!")
	}
	//@todo Transaction
	fmt.Println(string(wall.GetAddress()[:]))
	fmt.Println("Success!")
}

func Validate_Address(from, to string, amount uint64, nodeID string) bool {
	if !ValidateAddress(from) {
		fmt.Println("ERROR: Sender address is not valid")
		return false
	}
	if !ValidateAddress(to) {
		fmt.Println("ERROR: Recipient address is not valid")
		return false
	}
	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wall := wallets.GetWallet(from)
	if string(wall.GetAddress()[:]) != from {
		fmt.Println("Parameter verification failed!")
		return false
	}
	return true
}
