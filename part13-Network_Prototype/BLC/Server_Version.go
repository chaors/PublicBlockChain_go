package BLC

type Version struct {
	// 版本
	Version    int64
	// 当前节点区块的高度
	BestHeight int64
	//当前节点的地址
	AddrFrom   string
}
