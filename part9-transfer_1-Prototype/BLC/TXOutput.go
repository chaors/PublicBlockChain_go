package BLC

type TXOutput struct {
	//面值
	Value int64
	//暂时理解为用户名
	ScriptPubKey string
}

//验证当前交易输出属于某用户
func (txOutput *TXOutput) UnLockScriptPubKeyWithAddress(address string) bool {

	return txOutput.ScriptPubKey == address
}

