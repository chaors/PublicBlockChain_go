package BLC

import "crypto/sha256"

//默克尔树
type MerkleTree struct {
	//根节点
	RootNode *MerkleNode
}

//默克尔树节点
type MerkleNode struct {
	//做节点
	Left *MerkleNode
	//右节点
	Right *MerkleNode
	//节点数据
	Data []byte
}

//新建节点
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {

	mNode := MerkleNode{}

	if left == nil && right == nil {

		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	} else {

		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}

// 1 2 3 --> 1 2 3 3
//新建默克尔树
func NewMerkleTree(datas [][]byte) *MerkleTree {

	var nodes []*MerkleNode

	//如果是奇数，添加最后一个交易哈希拼凑为偶数个交易
	if len(datas) % 2 != 0 {

		datas = append(datas, datas[len(datas)-1])
	}

	//将每一个交易哈希构造为默克尔树节点
	for _, data := range datas {

		node := NewMerkleNode(nil, nil, data)
		nodes = append(nodes, node)
	}

	//将所有节点两两组合生成新节点，直到最后只有一个更节点
	for i := 0; i < len(datas)/2; i++ {

		var newLevel []*MerkleNode

		for j := 0; j < len(nodes); j += 2 {

			node := NewMerkleNode(nodes[j], nodes[j+1], nil)
			newLevel = append(newLevel, node)
		}

		nodes = newLevel
	}

	//取根节点返回
	mTree := MerkleTree{nodes[0]}

	return &mTree
}


