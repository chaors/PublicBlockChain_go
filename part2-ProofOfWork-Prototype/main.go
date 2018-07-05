/**
@author: chaors

@file:   main.go

@time:   2018/06/21 22:01

@desc:   区块信息的示例
*/


package main

import (
	"chaors.com/LearnGo/publicChaorsChain/part2-ProofOfWork-Prototype/BLC"
	"fmt"
)

func main() {

	//genesisBlock := BLC.CreateGenesisBlock("Genenis Block")
	//创建带有创世区块的区块链
	blockchain := BLC.CreateBlockchainWithGensisBlock()
	//添加一个新区快
	blockchain.AddBlockToBlockchain("first Block",
		blockchain.Blocks[len(blockchain.Blocks)-1].Height,
		blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockchain("second Block",
		blockchain.Blocks[len(blockchain.Blocks)-1].Height,
		blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockchain("third Block",
		blockchain.Blocks[len(blockchain.Blocks)-1].Height,
		blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	fmt.Println(blockchain)
}
