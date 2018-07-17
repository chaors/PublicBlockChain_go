package BLC

import "fmt"

func (cli *CLI) ResetUTXOSet(nodeID string)  {


	fmt.Println("TestMethod")

	blockchain := GetBlockchain(nodeID)
	defer blockchain.DB.Close()

	utxoSet := &UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()

	fmt.Println(blockchain.FindUTXOMap())
}
