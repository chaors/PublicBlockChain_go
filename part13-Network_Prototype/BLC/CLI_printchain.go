package BLC


//打印区块链
func (cli *CLI) printchain(nodeID string) {

	blockchain := GetBlockchain(nodeID)
	defer blockchain.DB.Close()

	blockchain.Printchain()
}
