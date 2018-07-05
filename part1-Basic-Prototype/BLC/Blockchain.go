/**
@author: chaors

@file:   Blockchain.go

@time:   2018/06/21 22:40

@desc:   区块链基础结构
*/


package BLC

type Blockchain struct {
	//有序区块的数组
	Blocks [] *Block
}

//1.创建带有创世区块的区块链
func CreateBlockchainWithGensisBlock() *Blockchain  {

	gensisBlock := CreateGenesisBlock("Gensis Block...")

	return &Blockchain{[] *Block{gensisBlock}}
}

//2.新增一个区块到区块链
func (blc *Blockchain) AddBlockToBlockchain(data string, height int64, prevHash []byte)  {

	//新建区块
	newBlock := NewBlock(data, height, prevHash)
	//上链
	blc.Blocks = append(blc.Blocks, newBlock)
}

