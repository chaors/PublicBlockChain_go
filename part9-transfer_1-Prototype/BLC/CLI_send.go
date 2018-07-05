package BLC



//转账
func (cli *CLI) send(from []string, to []string, amount []string)  {

	blockchain := GetBlockchain()
	defer blockchain.DB.Close()

	blockchain.MineNewBlock(from, to, amount)
}