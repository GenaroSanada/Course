package cli

import (
	"fmt"
	"Course/cli/wallet"
)

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := Wallet.NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}
