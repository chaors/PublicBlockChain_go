package BLC

type TXInput struct {
	//交易ID
	TxHash []byte
	//存储TXOutput在Vouts里的索引
	Vout int
	//数字签名
	ScriptSig string
}

//验证当前输入是否是当前地址的
func (txInput *TXInput) UnlockWithAddress(address string) bool  {

	return txInput.ScriptSig == address
}