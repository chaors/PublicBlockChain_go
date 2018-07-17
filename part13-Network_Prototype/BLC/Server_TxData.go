package BLC

// 同步中传递的交易类型
type TxData struct {
	// 节点地址
	AddFrom string
	// 交易
	TransactionBytes []byte
}
