package BLC



//新建区块链
func (cli *CLI)creatBlockchain(address string)  {

	blockchain := CreateBlockchainWithGensisBlock(address)
	defer blockchain.DB.Close()
}
