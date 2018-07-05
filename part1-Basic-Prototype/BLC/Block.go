/**
@author: chaors

@file:   Block.go

@time:   2018/06/21 21:46

@desc:   区块信息的基础结构
*/

package BLC

import (
	"time"
	"strconv"
	"fmt"
	"bytes"
	"crypto/sha256"
)

type Block struct {
	//1.区块高度
	Height int64
	//2.上一个区块HAsh
	PrevBlockHash []byte
	//3.交易数据
	Data []byte
	//4.时间戳
	Timestamp int64
	//5.Hash
	Hash []byte
}

//设置当前区块Hash
func (block *Block) SetHash() {

	//1.将高度，时间戳转换为字节数组
	//base:2  二进制形式
	heightBytes := IntToHex(block.Height)

	timeStampStr := strconv.FormatInt(block.Timestamp, 2)
	timeStamp := []byte(timeStampStr)

	//fmt.Println(heightBytes)
	//fmt.Println(timeStampStr)
	//fmt.Println(timeStamp)

	//2.拼接所有属性
	blockBytes := bytes.Join([][]byte{
		heightBytes,
		block.PrevBlockHash,
		block.Data,
		timeStamp,
		block.Hash}, []byte{})
	//fmt.Println(blockBytes)

	//3.将拼接后的字节数组转换为Hash值
	hash := sha256.Sum256(blockBytes)
	fmt.Println(hash)

	block.Hash = hash[:]
	//fmt.Println(block.Hash)

}

//1.创建新的区块
func NewBlock(data string, height int64, prevBlockHash []byte) *Block {

	//创建区块
	block := &Block{
		Height:        height,
		PrevBlockHash: prevBlockHash,
		Data:          []byte(data),
		Timestamp:     time.Now().Unix(),
		Hash:          nil}

	//设置HAsh值
	block.SetHash()

	return block
}

//单独方法生成创世区块
func CreateGenesisBlock(data string) *Block {

	return NewBlock(data, 1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

