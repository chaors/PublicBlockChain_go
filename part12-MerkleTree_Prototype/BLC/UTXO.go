package BLC

type UTXO struct {
	//来自交易的哈希
	TxHash []byte
	//在该交易VOuts里的下标
	Index int
	//未花费的交易输出
	Output *TXOutput
}

