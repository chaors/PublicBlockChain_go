package BLC

import "fmt"

func (cli *CLI)createWallet(nodeID string)  {

	wallets, _ := NewWallets(nodeID)
	wallets.CreateWallet(nodeID)

	fmt.Println(len(wallets.Wallets))
}
