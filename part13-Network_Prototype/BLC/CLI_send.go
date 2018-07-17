package BLC

import (
	"fmt"
	"strconv"
)

//转账
func (cli *CLI) send(from []string, to []string, amount []string, nodeID string, mineNow bool)  {

	blc := GetBlockchain(nodeID)
	defer blc.DB.Close()

	utxoSet := &UTXOSet{blc}

	// 由交易的第一个转账地址进行打包交易并挖矿
	if mineNow {

		blc.MineNewBlock(from, to, amount, nodeID)

		// 转账成功以后，需要更新UTXOSet
		utxoSet.Update()
	}else {

		// 把交易发送到矿工节点去进行验证
		fmt.Println("miner deal with the Tx...")

		// 遍历每一笔转账构造交易
		var txs []*Transaction
		nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
		for index, address := range from {

			value, _ := strconv.Atoi(amount[index])
			tx := NewTransaction(address, to[index], int64(value), utxoSet, txs, nodeID)
			txs = append(txs, tx)

			// 将交易发送给主节点
			sendTx(knowedNodes[0], tx)
		}
	}
}