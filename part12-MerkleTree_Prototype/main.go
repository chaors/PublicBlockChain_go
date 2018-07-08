package main

import (

	"chaors.com/LearnGo/publicChaorsChain/part11-transaction_1_Prototype/BLC"
)

func main() {

	cli := BLC.CLI{}
	cli.Run()

	//wallets := BLC.NewWallets()
	//address := hex.EncodeToString(wallet.GetAddress())
	//wallet1 := BLC.NewWallet()
	//address1 := hex.EncodeToString(wallet1.GetAddress())
	//
	//var from []string
	//var to []string
	//var amount []string
	//from = append(from, address)
	//to =append(to, address1)
	//amount = append(amount, "5")
	//
	//blc := BLC.CreateBlockchainWithGensisBlock(address)
	//blc.MineNewBlock(from, to, amount)

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
