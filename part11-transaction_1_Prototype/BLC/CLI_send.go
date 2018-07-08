package BLC



//转账
func (cli *CLI) send(from []string, to []string, amount []string)  {

	blockchain := GetBlockchain()
	defer blockchain.DB.Close()

	//打包交易并挖矿
	blockchain.MineNewBlock(from, to, amount)

	//转账成功以后，需要更新UTXOSet
	utxoSet := &UTXOSet{blockchain}
	utxoSet.Update()
}