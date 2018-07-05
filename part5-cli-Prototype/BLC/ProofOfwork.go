package BLC

import (
	"math/big"
	"bytes"
	"crypto/sha256"
	"fmt"
)

//期望计算的Hash值前面至少要有16个零
const targetBits = 16

type ProofOfWork struct {
	//求工作量的block
	Block *Block
	//工作量难度 big.Int大数存储
	target *big.Int
}

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
		fmt.Printf("\r%x", hash)
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


