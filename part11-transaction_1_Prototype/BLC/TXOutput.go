package BLC

import (
	"bytes"
)

type TXOutput struct {
	//面值
	Value int64
	//用户名
	Ripemd160Hash []byte  //用户名  公钥两次哈希后的值
}

func NewTXOutput(value int64,address string) *TXOutput {

	txOutput := &TXOutput{value,nil}

	// 设置Ripemd160Hash
	txOutput.Lock(address)

	return txOutput
}

//锁定
func (txOutput *TXOutput) Lock(address string) {

	version_pubKeyHash_checkSumBytes := Base58Decode([]byte(address))
	txOutput.Ripemd160Hash = version_pubKeyHash_checkSumBytes[1:len(version_pubKeyHash_checkSumBytes)-4]
}

//解锁
func (txOutput *TXOutput) UnLockScriptPubKeyWithAddress(address string) bool {

	version_pubKeyHash_checkSumBytes := Base58Decode([]byte(address))
	ripemd160Hash := version_pubKeyHash_checkSumBytes[1:len(version_pubKeyHash_checkSumBytes) - 4]

	//fmt.Println(txOutput.Ripemd160Hash, ripemd160Hash)
	return bytes.Compare(txOutput.Ripemd160Hash, ripemd160Hash) == 0
}


