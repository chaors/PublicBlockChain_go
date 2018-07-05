package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
)

//存储UTXO的表
const UTXOTableName  = "UTXOTableName"

type UTXOSet struct {
	
	Blc *Blockchain
}

//重置数据库表
func (utxoSet *UTXOSet) ResetUTXOSet()  {

	err := utxoSet.Blc.DB.Update(func(tx *bolt.Tx) error {

		//删除已经存在的UTXO表
		b := tx.Bucket([]byte(UTXOTableName))
		if b != nil{

			tx.DeleteBucket([]byte(UTXOTableName))

		}

		//新建UTXO表
		b, _ = tx.CreateBucket([]byte(UTXOTableName))
		if  b != nil{

			//找到区块链所有的UTXO
			txOutputMap := utxoSet.Blc.FindUTXOMAp()
			for keyHAsh , outs := range txOutputMap {

				txHash, _ := hex.DecodeString(keyHAsh)
				b.Put(txHash, outs.Se())
			}

		}

		return nil
	})
	if err != nil {

		log.Panic(err)
	}


}