package main

import (

	"chaors.com/LearnGo/publicChaorsChain/part9-transfer_1-Prototype/BLC"
)

func main() {

	cli := BLC.CLI{}
	cli.Run()

	//blc := BLC.CreateBlockchainWithGensisBlock("chaors")
	//utxos := blc.UnUTXOs("chaors")
	//fmt.Println(utxos)
}
