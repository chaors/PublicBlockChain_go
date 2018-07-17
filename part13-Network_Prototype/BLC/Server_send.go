package BLC

import (
	"fmt"
	"io"
	"bytes"
	"log"
	"net"
)

//COMMAND_VERSION
func sendVersion(toAddress string, blc *Blockchain)  {


	bestHeight := blc.GetBestHeight()
	payload := gobEncode(Version{NODE_VERSION, bestHeight, nodeAddress})

	request := append(commandToBytes(COMMAND_VERSION), payload...)

	sendData(toAddress, request)
}



//COMMAND_GETBLOCKS
func sendGetBlocks(toAddress string)  {

	payload := gobEncode(GetBlocks{nodeAddress})

	request := append(commandToBytes(COMMAND_GETBLOCKS), payload...)

	sendData(toAddress, request)

}

// 主节点将自己的所有的区块hash发送给钱包节点
//COMMAND_BLOCK
//
func sendInv(toAddress string, kind string, hashes [][]byte) {

	payload := gobEncode(Inv{nodeAddress,kind,hashes})

	request := append(commandToBytes(COMMAND_INV), payload...)

	sendData(toAddress, request)

}



func sendGetData(toAddress string, kind string ,blockHash []byte) {

	payload := gobEncode(GetData{nodeAddress,kind,blockHash})

	request := append(commandToBytes(COMMAND_GETDATA), payload...)

	sendData(toAddress, request)
}


func sendBlock(toAddress string, blockBytes []byte)  {


	payload := gobEncode(BlockData{nodeAddress,blockBytes})

	request := append(commandToBytes(COMMAND_BLOCK), payload...)

	sendData(toAddress, request)
}

func sendTx(toAddress string, tx *Transaction)  {

	data := TxData{nodeAddress, tx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes(COMMAND_TX), payload...)

	sendData(toAddress, request)
}

// 客户端向服务器发送数据
func sendData(to string, data []byte) {

	fmt.Printf("Client send message to server:%s...\n", to)

	conn, err := net.Dial("tcp", to)
	if err != nil {

		panic("error")
	}
	defer conn.Close()

	// 要发送的数据
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {

		log.Panic(err)
	}
}
