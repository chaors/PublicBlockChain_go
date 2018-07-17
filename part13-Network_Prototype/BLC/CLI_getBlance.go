package BLC

import "fmt"


//查询余额
func (cli *CLI) getBlance(address string, nodeID string) {

	fmt.Println("地址：" + address)

	blockchain := GetBlockchain(nodeID)
	defer blockchain.DB.Close()

	//amount := blockchain.GetBalance(address)  引入UTXOSet前的方法

	utxoSet := &UTXOSet{blockchain}
	amount := utxoSet.GetBalance(address)

	fmt.Printf("%s一共有%d个Token\n", address, amount)
}
