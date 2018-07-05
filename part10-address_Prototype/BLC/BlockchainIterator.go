package BLC

import (
	"github.com/boltdb/bolt"
	"log"
)

//区块链迭代器
type BlockchainIterator struct {
	//当前遍历hash
	CurrHash []byte
	//区块链数据库
	DB *bolt.DB
}

func (blcIterator *BlockchainIterator) Next() *Block {

	var block *Block

	err := blcIterator.DB.View(func(tx *bolt.Tx) error{

		b := tx.Bucket([]byte(blockTableName))

		if b != nil {
			currentBloclBytes := b.Get(blcIterator.CurrHash)

			// 获取到当前迭代器里面的currentHash所对应的区块
			block = DeSerializeBlock(currentBloclBytes)

			// 更新迭代器里面CurrentHash
			blcIterator.CurrHash = block.PrevBlockHash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return block
}

