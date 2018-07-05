/**
@author: chaors

@file:   Blockchain.go

@time:   2018/06/21 22:40

@desc:   区块链基础结构
*/

package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"fmt"
	"math/big"
	"time"
	"os"
	"encoding/hex"
	"strconv"
)

//相关数据库属性
const dbName = "chaorsBlockchain.db"
const blockTableName = "chaorsBlocks"
const newestBlockKey = "chNewestBlockKey"

type Blockchain struct {
	//最新区块的Hash
	Tip []byte
	//存储区块的数据库
	DB *bolt.DB
}

//1.创建创世区块
func CreateBlockchainWithGensisBlock(address string) *Blockchain {

	var blc *Blockchain

	//判断数据库是否存在
	if IsDBExists(dbName) {

		fmt.Println("创世区块已存在...")
		//os.Exit(1)

		//创建并打开数据库
		db, err := bolt.Open(dbName, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}

		var block *Block
		err = db.View(func(tx *bolt.Tx) error {

			b := tx.Bucket([]byte(blockTableName))
			if b != nil {

				hash := b.Get([]byte(newestBlockKey))
				blockBytes := b.Get(hash)
				block = DeSerializeBlock(blockBytes)
				fmt.Printf("\r######%d-%x\n", block.Nonce, hash)

				blc = &Blockchain{hash, db}
			}

			return nil
		})
		if err != nil {

			log.Panic(err)
		}

		return blc
		//os.Exit(1)
	}

	fmt.Println("正在创建创世区块...")

	//创建并打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {

		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucket([]byte(blockTableName))
		if err != nil {

			log.Panic(err)
		}

		if b != nil {

			//创币交易
			txCoinbase := NewCoinbaseTransaction(address)
			//创世区块
			gensisBlock := CreateGenesisBlock([]*Transaction{txCoinbase})
			//存入数据库
			err := b.Put(gensisBlock.Hash, gensisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			//存储最新区块hash
			err = b.Put([]byte(newestBlockKey), gensisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

			blc = &Blockchain{gensisBlock.Hash, db}
		}

		return nil
	})
	//更新数据库失败
	if err != nil {
		log.Fatal(err)
	}

	return blc
}

//2.新增一个区块到区块链 --> 包含交易的挖矿
func (blc *Blockchain) MineNewBlock(from []string, to []string, amount []string) {

	//send -from '["chaors"]' -to '["xyx"]' -amount '["5"]'

	//1.通过相关算法建立Transaction数组
	var txs []*Transaction

	//遍历输入输出，组装多笔交易
	for index, address := range from {

		value, _ := strconv.Atoi(amount[index])
		tx := NewTransaction(address, to[index], value, blc, txs)
		txs = append(txs, tx)
	}

	//2.挖矿
	//取上个区块的哈希和高度值
	var block *Block
	err := blc.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			hash := b.Get([]byte(newestBlockKey))
			blockBytes := b.Get(hash)
			block = DeSerializeBlock(blockBytes)
		}

		return nil
	})
	if err != nil {

		log.Panic(err)
	}

	//3.建立新区块
	block = NewBlock(txs, block.Height+1, block.Hash)

	//4.存储新区块
	err = blc.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			//fmt.Printf("444---%x\n\n", block.Txs[0].Vins[0].TxHash)
			//fmt.Println(block)

			err = b.Put(block.Hash, block.Serialize())
			if err != nil {

				log.Panic(err)
			}

			err = b.Put([]byte(newestBlockKey), block.Hash)
			if err != nil {

				log.Panic(err)
			}

			blc.Tip = block.Hash
		}

		return nil
	})
	if err != nil {

		log.Panic(err)
		//fmt.Print(err)
	}
}

//3.X 优化区块链遍历方法
func (blc *Blockchain) Printchain() {
	//迭代器
	blcIterator := blc.Iterator()

	//block := blcIterator.Next()
	//fmt.Printf("666---%x\n\n", block.Txs[0].Vins[0].txHash)

	for {

		block := blcIterator.Next()

		fmt.Println("------------------------------")
		fmt.Printf("Height：%d\n", block.Height)
		fmt.Printf("PrevBlockHash：%x\n", block.PrevBlockHash)
		fmt.Printf("Timestamp：%s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash：%x\n", block.Hash)
		fmt.Printf("Nonce：%d\n", block.Nonce)
		fmt.Println("Txs:")
		for _, tx := range block.Txs {

			fmt.Printf("%x\n", tx.TxHAsh)
			fmt.Println("Vins:")
			for _, in := range tx.Vins {
				fmt.Printf("txHash:%x\n", in.TxHash)
				fmt.Printf("Vout:%d\n", in.Vout)
				fmt.Printf("ScriptSig:%s\n\n", in.ScriptSig)
			}

			fmt.Println("Vouts:")
			for _, out := range tx.Vouts {
				fmt.Printf("Value:%d\n", out.Value)
				fmt.Printf("ScriptPubKey:%s\n\n", out.ScriptPubKey)
			}
		}
		fmt.Println("------------------------------\n\n")

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if big.NewInt(0).Cmp(&hashInt) == 0 {

			break
		}
	}
}

func (blc *Blockchain) Iterator() *BlockchainIterator {

	return &BlockchainIterator{blc.Tip, blc.DB}
}

//4.获取Blockchain对象
func GetBlockchain() *Blockchain {

	var blc *Blockchain
	//判断数据库是否存在
	if IsDBExists(dbName) {

		//创建并打开数据库
		db, err := bolt.Open(dbName, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}

		err = db.View(func(tx *bolt.Tx) error {

			b := tx.Bucket([]byte(blockTableName))
			if b != nil {

				hash := b.Get([]byte(newestBlockKey))
				blc = &Blockchain{hash, db}
			}

			return nil
		})
		if err != nil {

			log.Panic(err)
		}
	} else {

		fmt.Println("区块链不存在...")
		os.Exit(1)
	}

	return blc
}

//5.返回一个地址对应的UTXO的交易UTXOs
//func (blc *Blockchain) UnSpentTransactionsWithAddress(address string) []*Transaction {
func (blc *Blockchain) UTXOs(address string, txs []*Transaction) []*UTXO {

	//未花费的TXOutput
	var utxos []*UTXO

	//已经花费的TXOutput [hash:[]] [交易哈希：TxOutput对应的index]
	var spentTXOutputs = make(map[string][]int)

	//遍历器处理区块链上的UTXO
	blcIterator := blc.Iterator()
	for {

		block := blcIterator.Next()

		//fmt.Println(block)
		//fmt.Println()

		for _, tx := range block.Txs {

			// txHash

			// Vins
			//判断当前交易是否为创币交易
			if tx.IsCoinbaseTransaction() == false {

				for _, in := range tx.Vins {

					//验证当前输入是否是当前地址的
					if in.UnlockWithAddress(address) {

						key := hex.EncodeToString(in.TxHash)

						//fmt.Printf("lll%x\n", in.TxHash)
						//fmt.Println(key)
						spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
					}

				}
			}

			// Vouts
		Work:
			for index, out := range tx.Vouts {

				//验证当前输出是否是
				if out.UnLockScriptPubKeyWithAddress(address) {

					//fmt.Println(out)
					//fmt.Println(spentTXOutputs)

					//判断是否曾发生过交易
					if spentTXOutputs != nil {

						if len(spentTXOutputs) != 0 {

							//未花费UTXO标志
							isUnSpentUTXO := true

							//遍历spentTXOutputs
							for txHash, indexArray := range spentTXOutputs {

								//遍历TXOutputs下标数组
								for _, i := range indexArray {

									if index == i && txHash == hex.EncodeToString(tx.TxHAsh) {

										isUnSpentUTXO = false
										continue Work
									}
								}
							}

							if isUnSpentUTXO {

								utxo := &UTXO{tx.TxHAsh, index, out}
								utxos = append(utxos, utxo)
							}
						} else {

							utxo := &UTXO{tx.TxHAsh, index, out}
							utxos = append(utxos, utxo)
						}
					}
				}
			}
		}

		//找到创世区块，跳出循环
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		// Cmp compares x and y and returns:
		//
		//   -1 if x <  y
		//    0 if x == y
		//   +1 if x >  y
		if hashInt.Cmp(big.NewInt(0)) == 0 {

			break
		}
	}

	//处理未打包到区块链上的交易集里的UTXO
	for _, tx := range txs {

		if tx.IsCoinbaseTransaction() == false {
			for _, in := range tx.Vins {

				if in.UnlockWithAddress(address) {

					key := hex.EncodeToString(in.TxHash)

					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
				}
			}
		}
	}

	for _, tx := range txs {
	Work1:
		for index, out := range tx.Vouts {

			if out.UnLockScriptPubKeyWithAddress(address) {

				if len(spentTXOutputs) != 0 {

					for hash, indexArray := range spentTXOutputs {

						txHashStr := hex.EncodeToString(tx.TxHAsh)

						if hash == txHashStr {

							isUnSpentUTXO := true

							for _, outIndex := range indexArray {

								if index == outIndex {

									isUnSpentUTXO = false
									continue Work1
								}

								if isUnSpentUTXO {

									utxo := &UTXO{tx.TxHAsh, index, out}
									utxos = append(utxos, utxo)
								}
							}
						} else {

							utxo := &UTXO{tx.TxHAsh, index, out}
							utxos = append(utxos, utxo)
						}
					}
				} else {

					utxo := &UTXO{tx.TxHAsh, index, out}
					utxos = append(utxos, utxo)
				}
			}
		}
	}

	return utxos
}

//转账时查找可用的用于消费的UTXO  返回输入总金额和一个字典，UTXO集是一个字典类型，键是UTXO来源交易的哈希，值对该交易下UTXO对应TXOutput在Vounts中的下标
func (blc *Blockchain) FindSpendableUTXOs(address string, amount int, txs []*Transaction) (int64, map[string][]int) {

	//1.获取当前地址所有UTXO
	utxos := blc.UTXOs(address, txs)
	spendableUTXO := make(map[string][]int)

	//2.遍历UTXO
	//总的金额
	var value int64
	for _, utxo := range utxos {

		value += utxo.Output.Value
		txHash := hex.EncodeToString(utxo.TxHash)
		spendableUTXO[txHash] = append(spendableUTXO[txHash], utxo.Index)

		if value >= int64(amount) {

			break
		}
	}

	//余额不足
	if value < int64(amount) {

		fmt.Println("%s found.余额不足...", value)
		os.Exit(1)
	}

	return value, spendableUTXO
}

//查询余额
func (blc *Blockchain) GetBalance(address string) int64 {

	utxos := blc.UTXOs(address, []*Transaction{})

	var amount int64
	for _, out := range utxos {

		amount += out.Output.Value
	}

	return amount
}

//判断数据库是否存在
func IsDBExists(dbName string) bool {

	//if _, err := os.Stat(dbName); os.IsNotExist(err) {
	//
	//	return false
	//}

	_, err := os.Stat(dbName)
	if err == nil {

		return true
	}
	if os.IsNotExist(err) {

		return false
	}

	return true
}
