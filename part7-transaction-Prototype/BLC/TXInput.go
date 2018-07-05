package BLC

type TXInput struct {
	//交易ID
	TxHash []byte
	//存储TXOutput在Vout里的索引
	Vout int
	//数字签名
	ScriptSig string
}