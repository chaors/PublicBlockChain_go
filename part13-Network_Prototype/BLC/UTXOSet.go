package BLC

import (
	"log"
	"github.com/boltdb/bolt"
	"encoding/hex"
	"fmt"
	"os"
	"bytes"
)


//存储未花费交易输出的数据库表
const UTXOTableName  = "UTXOTableName"

type UTXOSet struct {

	Blockchain *Blockchain
}

// 1.重置数据库表
func (utxoSet *UTXOSet) ResetUTXOSet()  {

	//fmt.Println("resetUTXO:\n")
	//utxoSet.Blockchain.Printchain()

	err := utxoSet.Blockchain.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(UTXOTableName))

		// 删除原有UTXO表
		if b != nil {

			err := tx.DeleteBucket([]byte(UTXOTableName))
			if err!= nil {

				log.Panic(err)
			}
		}

		// 新建UTXO表
		b ,_ = tx.CreateBucket([]byte(UTXOTableName))
		if b != nil {

			//找到链上所有UTXO并存入数据库
			txOutputsMap := utxoSet.Blockchain.FindUTXOMap()

			for keyHash,outs := range txOutputsMap {

				TxHash,_ := hex.DecodeString(keyHash)

				b.Put(TxHash,outs.Serialize())

			}
		}

		return nil

	})
	if err != nil {

		log.Panic(err)
	}
}

// 2.查询某个地址的UTXO
func (utxoSet *UTXOSet) FindUTXOsForAddress(address string) []*UTXO {

	var utxos []*UTXO

	err := utxoSet.Blockchain.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(UTXOTableName))

		// 游标
		c := b.Cursor()
		for k, v := c.First(); k != nil; k,v = c.Next() {

			txOutputs := DeserializeTXOutputs(v)

			for _, utxo := range txOutputs.UTXOS {

				if utxo.Output.UnLockScriptPubKeyWithAddress(address) {

					utxos = append(utxos,utxo)
				}
			}
		}

		return nil
	})
	if err != nil {

		log.Panic(err)
	}

	return utxos
}

// 3.查询余额
func (utxoSet *UTXOSet) GetBalance(address string) int64 {

	UTXOS := utxoSet.FindUTXOsForAddress(address)

	var amount int64

	for _, utxo := range UTXOS  {

		amount += utxo.Output.Value
	}

	return amount
}

// 返回要凑多少钱，对应TXOutput的TX的Hash和index ???Set本身就是UTXO集合，里面的不全是未花费吗？？？？
func (utxoSet *UTXOSet) FindUnPackageSpendableUTXOS(address string, txs []*Transaction) []*UTXO {

	var unUTXOs []*UTXO
	spentTXOutputs := make(map[string][]int)

	for _,tx := range txs {

		if tx.IsCoinbaseTransaction() == false {

			for _, in := range tx.Vins {

				//是否能够解锁
				if in.UnlockWithAddress(address) {

					key := hex.EncodeToString(in.TxHash)
					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
				}
			}
		}
	}

	for _,tx := range txs {

	Work:
		for index,out := range tx.Vouts {

			if out.UnLockScriptPubKeyWithAddress(address) {

				if len(spentTXOutputs) != 0 {

					for hash,indexArray := range spentTXOutputs {

						TxHashStr := hex.EncodeToString(tx.TxHash)

						if hash == TxHashStr {

							var isUnSpent =true
							for _,outIndex := range indexArray {

								if index == outIndex {

									isUnSpent = false
									continue Work
								}

								if isUnSpent {

									utxo := &UTXO{tx.TxHash, index, out}
									unUTXOs = append(unUTXOs, utxo)
								}
							}
						} else {

							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				} else {

					utxo := &UTXO{tx.TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				}
			}
		}
	}

	return unUTXOs
}

//转账时查找可用的用于消费的UTXO组合
func (utxoSet *UTXOSet) FindSpendableUTXOs(address string,amount int64,txs []*Transaction) (int64,map[string][]int)  {

	unPackageUTXOS := utxoSet.FindUnPackageSpendableUTXOS(address, txs)

	spentableUTXO := make(map[string][]int)

	var value int64 = 0

	for _, UTXO := range unPackageUTXOS {

		value += UTXO.Output.Value
		TxHash := hex.EncodeToString(UTXO.TxHash)
		spentableUTXO[TxHash] = append(spentableUTXO[TxHash], UTXO.Index)

		if value >= amount{

			return  value, spentableUTXO
		}
	}

	// 钱还不够
	err := utxoSet.Blockchain.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(UTXOTableName))

		if b != nil {

			c := b.Cursor()
		UTXOBREAK:
			for k, v := c.First(); k != nil; k, v = c.Next() {

				txOutputs := DeserializeTXOutputs(v)

				for _, utxo := range txOutputs.UTXOS {

					value += utxo.Output.Value
					TxHash := hex.EncodeToString(utxo.TxHash)
					spentableUTXO[TxHash] = append(spentableUTXO[TxHash], utxo.Index)

					if value >= amount {

						break UTXOBREAK
					}
				}
			}
		}

		return nil
	})
	if err != nil {

		log.Panic(err)
	}

	if value < amount{

		fmt.Printf("%s found.余额不足...", value)
		os.Exit(1)
	}

	return  value, spentableUTXO
}

//更新UTXO 
func (utxoSet *UTXOSet) Update()  {

	// 1.找出最新区块
	block := utxoSet.Blockchain.Iterator().Next()

	// 未花费的UTXO  键为对应交易哈希，值为TXOutput数组
	outsMap := make(map[string] *TXOutputs)
	// 新区快的交易输入,这些交易输入引用的TXOutput被消耗，应该从UTXOSet删除
	ins := []*TXInput{}

	// 2.遍历区块交易找出交易输入
	for _, tx := range block.Txs {

		//遍历交易输入，
		for _, in := range tx.Vins {

			ins = append(ins, in)
		}
	}

	// 2.遍历交易输出
	for _, tx := range block.Txs {

		utxos := []*UTXO{}

		for index, out := range tx.Vouts {

			//未花费标志
			isUnSpent := true
			for _, in := range ins {

				if in.Vout == index && bytes.Compare(tx.TxHash, in.TxHash) == 0 &&
					bytes.Compare(out.Ripemd160Hash, Ripemd160Hash(in.PublicKey)) == 0 {

						isUnSpent = false
						continue
				}
			}

			if isUnSpent {

				utxo := &UTXO{tx.TxHash,index,out}
				utxos = append(utxos,utxo)
			}
		}

		if len(utxos) > 0 {

			TxHash := hex.EncodeToString(tx.TxHash)
			outsMap[TxHash] = &TXOutputs{utxos}
		}
	}

	//3. 删除已消耗的TXOutput
	err := utxoSet.Blockchain.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(UTXOTableName))
		if b != nil {

			for _, in := range ins {

				txOutputsBytes := b.Get(in.TxHash)

				//如果该交易输入无引用的交易哈希
				if len(txOutputsBytes) == 0 {

					continue
				}
				txOutputs := DeserializeTXOutputs(txOutputsBytes)

				// 判断是否需要
				isNeedDelete := false

				//缓存来自该交易还未花费的UTXO
				utxos := []*UTXO{}

				for _, utxo := range txOutputs.UTXOS {

					if in.Vout == utxo.Index && bytes.Compare(utxo.Output.Ripemd160Hash, Ripemd160Hash(in.PublicKey)) == 0 {

						isNeedDelete = true
					}else {

						//txOutputs中剩余未花费的txOutput
						utxos = append(utxos,utxo)
					}
				}

				if isNeedDelete {

					b.Delete(in.TxHash)

					if len(utxos) > 0 {

						preTXOutputs := outsMap[hex.EncodeToString(in.TxHash)]
						preTXOutputs.UTXOS = append(preTXOutputs.UTXOS, utxos...)
						outsMap[hex.EncodeToString(in.TxHash)] = preTXOutputs
					}
				}
			}

			// 4.新增交易输出到UTXOSet
			for keyHash, outPuts := range outsMap {

				keyHashBytes, _ := hex.DecodeString(keyHash)
				b.Put(keyHashBytes, outPuts.Serialize())
			}
		}

		return nil
	})
	if err != nil{

		log.Panic(err)
	}
}