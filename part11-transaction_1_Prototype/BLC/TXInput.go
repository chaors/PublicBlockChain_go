package BLC

import "bytes"

type TXInput struct {
	//交易ID
	TxHash []byte
	//存储TXOutput在Vouts里的索引
	Vout int
	//数字签名
	Signature []byte
	//公钥
	PublicKey []byte
}

//验证当前输入是否是当前地址的
func (txInput *TXInput) UnlockWithAddress(address string) bool  {

	//base58解码
	version_pubKeyHash_checkSumBytes := Base58Decode([]byte(address))
	//去除版本得到地反编码的公钥两次哈希后的值
	ripemd160Hash := version_pubKeyHash_checkSumBytes[1:len(version_pubKeyHash_checkSumBytes)-4]

	//Ripemd160Hash算法得到公钥两次哈希后的值
	ripemd160HashNew := Ripemd160Hash(txInput.PublicKey)

	return bytes.Compare(ripemd160HashNew, ripemd160Hash) == 0
}