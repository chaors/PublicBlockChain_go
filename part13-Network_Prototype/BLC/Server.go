package BLC

import (
	"net"
	"fmt"
	"log"
	"io/ioutil"
)


func StartServer(nodeID string, minerAdd string) {

	// 当前节点IP地址
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	// 挖矿节点设置
	if len(minerAdd) > 0 {

		miningAddress = minerAdd
	}

	// 启动网络监听服务
	ln, err := net.Listen(PROTOCOL, nodeAddress)
	if err != nil {

		log.Panic(err)
	}
	defer ln.Close()

	blc := GetBlockchain(nodeID)
	//fmt.Println("startserver\n")
	//blc.Printchain()

	// 第一个终端：端口为3000,启动的就是主节点
	// 第二个终端：端口为3001，钱包节点
	// 第三个终端：端口号为3002，矿工节点
	if nodeAddress != knowedNodes[0] {

		// 该节点不是主节点，钱包节点向主节点请求数据
		sendVersion(knowedNodes[0], blc)
	}

	for {

		// 接收客户端发来的数据
		connc, err := ln.Accept()
		if err != nil {

			log.Panic(err)
		}

		// 不同的命令采取不同的处理方式
		go handleConnection(connc, blc)
	}
}

// 客户端命令处理器
func handleConnection(conn net.Conn, blc *Blockchain) {

	//fmt.Println("handleConnection:\n")
	//blc.Printchain()

	// 读取客户端发送过来的所有的数据
	request, err := ioutil.ReadAll(conn)
	if err != nil {

		log.Panic(err)
	}

	fmt.Printf("\nReceive a Message:%s\n", request[:COMMANDLENGTH])

	command := bytesToCommand(request[:COMMANDLENGTH])

	switch command {

	case COMMAND_VERSION:
		handleVersion(request, blc)

	case COMMAND_ADDR:
		handleAddr(request, blc)

	case COMMAND_BLOCK:
		handleBlock(request, blc)

	case COMMAND_GETBLOCKS:
		handleGetblocks(request, blc)

	case COMMAND_GETDATA:
		handleGetData(request, blc)

	case COMMAND_INV:
		handleInv(request, blc)

	case COMMAND_TX:
		handleTx(request, blc)

	default:
		fmt.Println("Unknown command!")
	}

	defer conn.Close()
}

// 节点是否在已知节点中
func nodeIsKnown(addr string) bool {

	for _, node := range knowedNodes {

		if node == addr {

			return true
		}
	}

	return false
}
