/**
@author: chaors

@file:   main.go

@time:   2018/06/21 22:01

@desc:   boltdb存储区块信心
*/


package main

import (
	"chaors.com/LearnGo/publicChaorsChain/part3-boltdb-Prototype/BLC"
	"github.com/boltdb/bolt"
	"log"
	"fmt"
)

func main() {

	blockchain := BLC.CreateBlockchainWithGensisBlock()
	//添加一个新区快
	blockchain.AddBlockToBlockchain("first Block",
		blockchain.Blocks[len(blockchain.Blocks)-1].Height,
		blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockchain("second Block",
		blockchain.Blocks[len(blockchain.Blocks)-1].Height,
		blockchain.Blocks[len(blockchain.Blocks)-1].Hash)

	//1.数据库创建
	//在这里gland直接运行，生成的my.db在main.go上层目录;命令行build在运行的话是当前目录！！！
	db, err := bolt.Open("chaorsBlock.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//2.创建表
	err = db.Update(func(tx *bolt.Tx) error {

		//创建叫"MyBucket"的表
		b := tx.Bucket([]byte("MyBlocks"))
		if b == nil {

			_, err := tx.CreateBucket([]byte("MyBlocks"))
			if err != nil {
				log.Fatal(err)
			}
		}


		//一定要返回nil
		return nil
	})

	//更新数据库失败
	if err != nil {
		log.Fatal(err)
	}


	//3.更新表数据
	err = db.Update(func(tx *bolt.Tx) error {

		//取出叫"MyBucket"的表
		b := tx.Bucket([]byte("MyBlocks"))

		//往表里面存储数据
		if b != nil {

			err := b.Put(blockchain.Blocks[0].Hash, blockchain.Blocks[0].Serialize())
			err = b.Put(blockchain.Blocks[1].Hash, blockchain.Blocks[1].Serialize())
			err = b.Put(blockchain.Blocks[1].Hash, blockchain.Blocks[2].Serialize())
			err = b.Put([]byte("headBlock"), blockchain.Blocks[2].Serialize())
			if err != nil {
				log.Fatal(err)
			}
		}

		//一定要返回nil
		return nil
	})

	//更新数据库失败
	if err != nil {
		log.Fatal(err)
	}


	//4.查看表数据
	err = db.View(func(tx *bolt.Tx) error {

		//取出叫"MyBucket"的表
		b := tx.Bucket([]byte("MyBlocks"))

		//往表里面存储数据
		if b != nil {

			data := b.Get(blockchain.Blocks[0].Hash)
			fmt.Printf("%v \n", BLC.DeSerializeBlock(data))
			data = b.Get(blockchain.Blocks[1].Hash)
			fmt.Printf("%s\n", data)
		}

		//一定要返回nil
		return nil
	})

	//更新数据库失败
	if err != nil {
		log.Fatal(err)
	}



}
