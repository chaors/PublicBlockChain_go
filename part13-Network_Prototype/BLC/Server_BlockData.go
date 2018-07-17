package BLC

// 用于节点间发送一个区块
type BlockData struct {
	// 节点地址
	AddrFrom string
	// 序列化区块
	BlockBytes []byte
}
