package BLC

import "fmt"

func (cli *CLI) getAddressList(nodeID string)  {

	fmt.Println("All addresses:")

	wallets, _ := NewWallets(nodeID)
	for address, _ := range wallets.Wallets {

		fmt.Println(address)
	}
}
