package BLC

import "fmt"


//查询余额
func (cli *CLI) getBlance(address string) {

	fmt.Println("地址：" + address)

	blockchain := GetBlockchain()
	defer blockchain.DB.Close()

	amount := blockchain.GetBalance(address)

	fmt.Printf("%s一共有%d个Token\n", address, amount)

}
