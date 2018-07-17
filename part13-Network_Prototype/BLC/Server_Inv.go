package BLC

// 向其他节点展示自己拥有的区块和交易
type Inv struct {
	// 自己的地址
	AddrFrom string
	// 类型 block tx
	Type     string
	// hash二维数组
	Items    [][]byte
}
