package main

import (

	"chaors.com/LearnGo/publicChaorsChain/part10-address_Prototype/BLC"
)

func main() {

	cli := BLC.CLI{}
	cli.Run()

	/**
	//Wallet
	ripemd160.New()
	wallet := BLC.NewWallet()
	address := wallet.GetAddress()

	fmt.Printf("%s\n", address)
	//1CiS8axkfLGQUYaeZsuS2Fpv4nVcd6HQqk 和比特币地址相同，可以blockchaininfo查询余额
	//当然一定为0

	//判断地址有效性
	fmt.Println(wallet.IsValidForAddress(address))
	//修改address
	fmt.Println(wallet.IsValidForAddress([]byte("1CiS8axkfLGQUYaeZsuS2Fpv4nVcd6HQqkk")))

	//Wallets
	wallets := BLC.NewWallets()
	wallets.CreateWallet()
	wallets.CreateWallet()
	wallets.CreateWallet()
	fmt.Println()
	fmt.Println(wallets)

	//blc := BLC.CreateBlockchainWithGensisBlock("chaors")
	//utxos := blc.UnUTXOs("chaors")
	//fmt.Println(utxos)
	*/
}
