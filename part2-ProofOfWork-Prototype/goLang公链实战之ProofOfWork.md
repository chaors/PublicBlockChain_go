# goLang公链实战之ProofOfWork

今天来实现工作量证明。

### 区块新属性Nonce
我们先来看一下上节实现的区块结构：
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

我们都知道一个合法区块的诞生其哈希值必须满足指定的条件，比特币采用的是工作量证明。我们这里用go开发的公链也采用POW一致性算法来产生合法性区块。

因此，区块必须不断产生哈希直到满足POW的哈希值产生才能添加到主链上成为合法区块。看看上图区块的基本结构，对于一个区块来说，1-4项属性都市固定，而区块哈希又是由这些属性拼接生成的。所以，要想让区块哈希能不断变化，必须引入一个变量Nonce。

引入Nonce后，就可以通过改变Nonce值来不断产生新的哈希值直到找到满足条件的哈希。

### POW难度targetBits

前面用Python简单介绍过[区块链中的挖矿概念](https://www.jianshu.com/p/b39c361687c1),一般地，对于256位的哈希值来说设定挖矿条件的方式往往是：前多少位为0。targetBits便是用于指定目标哈希需满足的条件的，即计算的哈希值必须前targetBits位为0.

而对于一个256位的二进制串判断前多少位为0显得很繁琐，我们可以巧妙地通过位移运算将这一判断转换为一个数学问题。eg:

假设哈希值得位数为8，当前targetBits为2(256位亦然，用8举例是位数少便于描述)，那么目标哈希值必须满足前两位都是0。从临界情况入手，当第2位不为0时的最小的数为0100 0000，只要小于这个数就是符合条件的哈希值。那么这些数是怎么找到的呢？

1.8位哈希值最小的非0值为0000 0001
2.该值左移8-targetBits位，0000 0001 << 6 = 0100 0000 = target
3.if hash < target 区块合法

### Block结构完善
##### Nonce
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
	//6.Nonce  符合工作量证明的随机数
	Nonce int64
}
```
##### 新区块产生
```
//1.创建新的区块
func NewBlock(data string, height int64, prevBlockHash []byte) *Block {

	//创建区块
	block := &Block{
		Height:        height,
		PrevBlockHash: prevBlockHash,
		Data:          []byte(data),
		Timestamp:     time.Now().Unix(),
		Hash:          nil,
		Nonce:         0}

	//调用工作量证明返回有效的Hash
	pow := NewProofOfWork(block)
	hash, nonce := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	fmt.Printf("\r%d-%x\n", nonce, hash)

	return block
}
```

### ProofOfWork
##### 基本结构
```
//期望计算的Hash值前面至少要有16个零
const targetBits = 16

type ProofOfWork struct {
	//求工作量的block
	Block *Block
	//工作量难度 big.Int大数存储
	target *big.Int
}
```

##### 创建新的POW对象
```
//创建新的工作量证明对象
func NewProofOfWork(block *Block) *ProofOfWork {

	/**
	target计算方式  假设：Hash为8位，targetBit为2位
	eg:0000 0001(8位的Hash)
	1.8-2 = 6 将上值左移6位
	2.0000 0001 << 6 = 0100 0000 = target
	3.只要计算的Hash满足 ：hash < target，便是符合POW的哈希值
	*/

	//1.创建一个初始值为1的target
	target := big.NewInt(1)
	//2.左移bits(Hash) - targetBit 位
	target = target.Lsh(target, 256-targetBits)

	return &ProofOfWork{block, target}
}
```

##### 哈希值的预选值
```
//拼接区块属性，返回字节数组
func (pow *ProofOfWork) prepareData(nonce int) []byte {

	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,
			pow.Block.Data,
			IntToHex(pow.Block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
			IntToHex(int64(pow.Block.Height)),
		},
		[]byte{},
	)

	return data
}
```

##### 区块有效性验证
```
//判断当前区块是否有效
func (proofOfWork *ProofOfWork) IsValid() bool  {

	//比较当前区块哈希值与目标哈希值
	var hashInt big.Int
	hashInt.SetBytes(proofOfWork.Block.Hash)

	if proofOfWork.target.Cmp(&hashInt) == 1 {

		return true
	}

	return false
}
```

##### 挖矿(产生有效的哈希值)
```
//运行工作量证明
func (proofOfWork *ProofOfWork) Run() ([]byte, int64) {

	//1.将Block属性拼接成字节数组

	//2.生成hash
	//3.判断Hash值有效性，如果满足条件跳出循环

	//用于寻找目标hash值的随机数
	nonce := 0
	//存储新生成的Hash值
	var hashInt big.Int
	var hash [32]byte

	for {
		//准备数据
		dataBytes := proofOfWork.prepareData(nonce)
		//生成Hash
		hash = sha256.Sum256(dataBytes)

		//\r将当前打印行覆盖
		//fmt.Printf("\r%x", hash)
		//存储Hash到hashInt
		hashInt.SetBytes(hash[:])
		//验证Hash
		if proofOfWork.target.Cmp(&hashInt) == 1 {

			break
		}
		nonce++
	}

	return hash[:], int64(nonce)
}
```

### POW测试
```

package main

import (
	"chaors.com/LearnGo/publicChaorsChain/part2-ProofOfWork-Prototype/BLC"
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

运行后，consle会不断打印计算出的Hash值，直到计算出的Hash值满足条件。当我们调整POW难度targetBits，会发现计算的时间有所改变。targetBits越大，挖矿难度越大，新区快产生的时间也就越久。










