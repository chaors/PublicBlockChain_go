package BLC

import "fmt"

func (cli *CLI) getAddressList()  {

	fmt.Println("All addresses:")

	wallets, _ := NewWallets()
	for address, _ := range wallets.Wallets {

		fmt.Println(address)
	}
}
