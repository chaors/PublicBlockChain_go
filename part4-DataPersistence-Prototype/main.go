/**
@author: chaors

@file:   main.go

@time:   2018/06/21 22:01

@desc:   boltdb存储区块信心
*/

package main

import (
	"chaors.com/LearnGo/publicChaorsChain/part4-DataPersistence-Prototype/BLC"
)

func main() {

	blockchain := BLC.CreateBlockchainWithGensisBlock()
	defer blockchain.DB.Close()

	//添加一个新区快
	blockchain.AddBlockToBlockchain("4th Block")
	blockchain.AddBlockToBlockchain("5th Block")
	//blockchain.AddBlockToBlockchain("third Block")

	blockchain.Printchain()

}
