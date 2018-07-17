package BLC

//新建区块链
func (cli *CLI)creatBlockchain(address string, nodeID string)  {

	blockchain := CreateBlockchainWithGensisBlock(address, nodeID)
	defer blockchain.DB.Close()
}
