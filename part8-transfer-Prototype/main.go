package main

import (

	"chaors.com/LearnGo/publicChaorsChain/part8-transfer-Prototype/BLC"
	//"fmt"
	//"os"
	//"flag"
	//"log"
	//"fmt"
)

func main() {

	cli := BLC.CLI{}
	cli.Run()

	//blc := BLC.CreateBlockchainWithGensisBlock("chaors")
	//utxos := blc.UnUTXOs("chaors")
	//fmt.Println(utxos)
}
