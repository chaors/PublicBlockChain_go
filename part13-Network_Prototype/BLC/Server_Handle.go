package BLC

import (
	"log"
	"encoding/gob"
	"bytes"
	"fmt"
	"encoding/hex"
	"github.com/boltdb/bolt"
)

// Version命令处理器
func handleVersion(request []byte, blc *Blockchain)  {

	var buff bytes.Buffer
	var payload Version

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	// 提取最大区块高度作比较
	bestHeight := blc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if bestHeight > foreignerBestHeight {

		// 向请求节点回复自身Version信息
		sendVersion(payload.AddrFrom, blc)
	} else if bestHeight < foreignerBestHeight {

		// 向请求节点要信息
		sendGetBlocks(payload.AddrFrom)
	}

	// 添加到已知节点中
	if !nodeIsKnown(payload.AddrFrom) {

		knowedNodes = append(knowedNodes, payload.AddrFrom)
	}
}

func handleAddr(request []byte, blc *Blockchain)  {




}

func handleGetblocks(request []byte, blc *Blockchain)  {

	var buff bytes.Buffer
	var payload GetBlocks

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := blc.GetBlockHashes()

	sendInv(payload.AddrFrom, BLOCK_TYPE, blocks)
}

func handleGetData(request []byte, blc *Blockchain)  {

	var buff bytes.Buffer
	var payload GetData

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	if payload.Type == BLOCK_TYPE {

		block, err := blc.GetBlock([]byte(payload.Hash))
		if err != nil {

			return
		}

		sendBlock(payload.AddrFrom, block)
	}

	if payload.Type == TX_TYPE {

		// 取出交易
		TxHash := hex.EncodeToString(payload.Hash)
		tx := memTxPool[TxHash]

		sendTx(payload.AddrFrom, &tx)
	}
}

func handleBlock(request []byte, blc *Blockchain)  {

	//fmt.Println("handleblock:\n")
	//blc.Printchain()

	var buff bytes.Buffer
	var payload BlockData

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	block := DeSerializeBlock(payload.BlockBytes)
	if block == nil {

		fmt.Printf("Block nil")
	}

	err = blc.AddBlock(block)
	if err != nil {

		log.Panic(err)
	}
	fmt.Printf("add block %x succ.\n", block.Hash)
	//blc.Printchain()

	if len(unslovedHashes) > 0 {

		sendGetData(payload.AddrFrom, BLOCK_TYPE, unslovedHashes[0])
		unslovedHashes = unslovedHashes[1:]
	}else {

		//blc.Printchain()
		utxoSet := &UTXOSet{blc}
		utxoSet.ResetUTXOSet()
	}
}


func handleTx(request []byte, blc *Blockchain)  {

	var buff bytes.Buffer
	var payload TxData

	dataBytes := request[COMMANDLENGTH:]
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	tx := DeserializeTransaction(payload.TransactionBytes)
	memTxPool[hex.EncodeToString(tx.TxHash)] = tx

	// 自身为主节点，需要将交易转发给矿工节点
	if nodeAddress == knowedNodes[0] {

		for _, node := range knowedNodes {

			if node != nodeAddress && node != payload.AddFrom {

				sendInv(node, TX_TYPE, [][]byte{tx.TxHash})
			}
		}
	} else {

		//fmt.Println(len(memTxPool), len(miningAddress))
		if len(memTxPool) >= minMinerTxCount && len(miningAddress) > 0 {

		MineTransactions:

			var txs []*Transaction
			// 创币交易，作为挖矿奖励
			coinbaseTx := NewCoinbaseTransaction(miningAddress)
			txs = append(txs, coinbaseTx)

			var verifyTxs []*Transaction

			for id := range memTxPool {

				tx := memTxPool[id]
				if blc.VerifyTransaction(&tx, verifyTxs) {

					txs = append(txs, &tx)
					verifyTxs = append(verifyTxs, &tx)
				}else {

					log.Panic("the transaction  invalid...\n")
				}
			}

			fmt.Println("All transactions verified succ!\n")

			// 建立新区块
			var block *Block
			// 取出上一个区块
			err = blc.DB.View(func(tx *bolt.Tx) error {

				b := tx.Bucket([]byte(blockTableName))
				if b != nil {

					hash := b.Get([]byte(newestBlockKey))
					block = DeSerializeBlock(b.Get(hash))
				}

				return nil
			})
			if err != nil {

				log.Panic(err)
			}

			//构造新区块
			block = NewBlock(txs, block.Height+1, block.Hash)

			fmt.Println("New block is mined!")

			// 添加到数据库
			err = blc.DB.Update(func(tx *bolt.Tx) error {

				b := tx.Bucket([]byte(blockTableName))
				if b != nil {

					b.Put(block.Hash, block.Serialize())
					b.Put([]byte(newestBlockKey), block.Hash)
					blc.Tip = block.Hash

				}
				return nil
			})
			if err != nil {

				log.Panic(err)
			}

			utxoSet := UTXOSet{blc}
			//utxoSet.Update()
			utxoSet.ResetUTXOSet()

			// 去除内存池中打包到区块的交易
			for _, tx := range txs {

				fmt.Println("delete...")
				TxHash := hex.EncodeToString(tx.TxHash)
				delete(memTxPool, TxHash)
			}

			// 发送区块给其他节点
			//sendBlock(knowedNodes[0], block.Serialize())
			for _, node := range knowedNodes {

				if node != nodeAddress {

					sendBlock(node, block.Serialize())
				}
			}

			if len(memTxPool) > 0 {

				goto MineTransactions
			}
		}
	}
}


func handleInv(request []byte, blc *Blockchain)  {

	var buff bytes.Buffer
	var payload Inv

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// Ivn 3000 block hashes [][]
	if payload.Type == BLOCK_TYPE {

		fmt.Println(payload.Items)

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, BLOCK_TYPE , blockHash)

		if len(payload.Items) >= 1 {

			unslovedHashes = payload.Items[1:]
		}
	}

	if payload.Type == TX_TYPE {

		TxHash := payload.Items[0]

		// 添加到交易池
		if memTxPool[hex.EncodeToString(TxHash)].TxHash == nil {

			sendGetData(payload.AddrFrom, TX_TYPE, TxHash)
		}
	}
}