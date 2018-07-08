package BLC


//打印区块链
func (cli *CLI) printchain() {

	blockchain := GetBlockchain()
	defer blockchain.DB.Close()

	blockchain.Printchain()
}
