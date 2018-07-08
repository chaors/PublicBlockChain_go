# goLang公链实战之区块链基础结构

堕落了一段时间，终于又找回了学习的动力，满血归来。。。

我们知道在如火如荼的区块链应用红海，goLang越来越多地发挥着不可替代的作用。一方面取决于其语法的简单性，一方面其具备C++高效处理的特性。今天，我们就用go语言开始构建一个简单但是具备区块链完整功能的公链项目。

由于之前已经用[Python构建过简单的区块链结构](https://www.jianshu.com/p/ecfb2a9040a3)，所以对区块基本结构的东西不再做详细赘述。

#废话少说上干货

##区块Block

####Block
```
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
```

####设置当前区块Hash
```
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
```

####创建新区快
```
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
```

####创世区块创建
```
func CreateGenesisBlock(data string) *Block {

	return NewBlock(data, 1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}
```

###区块链BlockChain

####区块链
```
type Blockchain struct {
	//有序区块的数组
	Blocks [] *Block
}
```

####创建带有创世区块的区块链
```
func CreateBlockchainWithGensisBlock() *Blockchain  {

	gensisBlock := CreateGenesisBlock("Gensis Block...")

	return &Blockchain{[] *Block{gensisBlock}}
}
```

####新增一个区块到区块链
```
func (blc *Blockchain) AddBlockToBlockchain(data string, height int64, prevHash []byte)  {

	//新建区块
	newBlock := NewBlock(data, height, prevHash)
	//上链
	blc.Blocks = append(blc.Blocks, newBlock)
}
```

###utils辅助工具(int64转化为byte数组)
```

import (
	"bytes"
	"encoding/binary"
	"log"
)

//将int64转换为bytes
func IntToHex(num int64) []byte  {

	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {

		log.Panic(err)
	}

	return buff.Bytes()
}
```

###测试Demo
```
/**
@author: chaors

@file:   main.go

@time:   2018/06/21 22:01

@desc:   区块信息的示例
*/

package main

import (
	"chaors.com/LearnGo/publicChaorsChain/part1-Basic-Prototype/BLC"
	"fmt"
)

func main() {

	//genesisBlock := BLC.CreateGenesisBlock("Genenis Block")
	//创建带有创世区块的区块链
	blockchain := BLC.CreateBlockchainWithGensisBlock()
	//添加一个新区快
	blockchain.AddBlockToBlockchain("first Block",
		blockchain.Blocks[len(blockchain.Blocks)-1].Height,
		blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockchain("second Block",
		blockchain.Blocks[len(blockchain.Blocks)-1].Height,
		blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	blockchain.AddBlockToBlockchain("third Block",
		blockchain.Blocks[len(blockchain.Blocks)-1].Height,
		blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	fmt.Println(blockchain)
}
```

###运行结果

我们看到运行的结果，打印的内容为包含创世区块在内的四个区块的区块链。

![image.png](https://upload-images.jianshu.io/upload_images/830585-9a952234ba83dd97.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)






