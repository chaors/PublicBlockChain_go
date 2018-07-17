/**
@author: chaors

@file:   Block.go

@time:   2018/06/21 21:46

@desc:   区块信息的基础结构
*/

package BLC

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
	"fmt"
)

type Block struct {
	//1.区块高度
	Height int64
	//2.上一个区块HAsh
	PrevBlockHash []byte
	//3.交易数据
	Txs [] *Transaction
	//4.时间戳
	Timestamp int64
	//5.Hash
	Hash []byte
	//6.Nonce  符合工作量证明的随机数
	Nonce int64
}

//区块序列化
func (block *Block) Serialize() []byte  {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

//区块反序列化
func DeSerializeBlock(blockBytes []byte) *Block  {

	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))

	err := decoder.Decode(&block)

	if err != nil {

		log.Panic(err)
	}

	return &block
}


//1.创建新的区块
func NewBlock(txs []*Transaction, height int64, prevBlockHash []byte) *Block {

	//创建区块
	block := &Block{
		Height:        height,
		PrevBlockHash: prevBlockHash,
		Txs:           txs,
		Timestamp:     time.Now().Unix(),
		Hash:          nil,
		Nonce:         0}

	//调用工作量证明返回有效的Hash
	pow := NewProofOfWork(block)
	hash, nonce := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	fmt.Printf("\r######%d-%x\n", nonce, hash)

	return block
}

//单独方法生成创世区块
func CreateGenesisBlock(txs []*Transaction) *Block {

	return NewBlock(
		txs,
		1,
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		)
}

// 需要将Txs转换成[]byte
func (block *Block) HashTransactions() []byte  {

	//引入MerkleTree前的交易哈希
	//var TxHashes [][]byte
	//var TxHash [32]byte
	//
	//for _, tx := range block.Txs {
	//
	//	TxHashes = append(TxHashes, tx.TxHash)
	//}
	//TxHash = sha256.Sum256(bytes.Join(TxHashes, []byte{}))
	//
	//return TxHash[:]

	//默克尔树根节点表示交易哈希
	var transactions [][]byte

	for _, tx := range block.Txs {

		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}
