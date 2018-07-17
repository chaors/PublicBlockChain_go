package BLC

// 用于请求区块或交易
type GetData struct {
	// 节点地址
	AddrFrom string
	// 请求类型  是block还是tx
	Type     string
	// 区块哈希或交易哈希
	Hash       []byte
}
