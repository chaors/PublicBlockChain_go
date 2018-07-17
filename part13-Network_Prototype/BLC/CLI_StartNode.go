package BLC

import (
	"fmt"
	"os"
)

func (cli *CLI) startNode(nodeID string, minerAdd string)  {

	fmt.Printf("start Server:localhost:%s\n", nodeID)
	// 挖矿地址判断
	if len(minerAdd) > 0 {

		if IsValidForAddress([]byte(minerAdd)) {

			fmt.Printf("Miner:%s is ready to mining...\n", minerAdd)
		}else {

			fmt.Println("Server address invalid....\n")
			os.Exit(0)
		}
	}

	// 启动服务器
	StartServer(nodeID, minerAdd)
}