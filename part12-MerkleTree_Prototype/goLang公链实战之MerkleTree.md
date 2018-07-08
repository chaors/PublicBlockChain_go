# goLang公链实战之Merkle树

# MerkleTree
MerkleTree，通常也被称作Hash Tree，顾名思义，就是存储hash值的一棵树。Merkle树的叶子是数据块(例如，文件或者文件的集合)的hash值。非叶节点是其对应子节点串联字符串的hash。

下面是一幅表示MerkleTree结构的图：

![MerkleTree结构.png](https://upload-images.jianshu.io/upload_images/830585-1c6496ed8a8efd42.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

那么MerkleTree对我们构造公链有什么用呢？

我们知道完整的比特币数据库已达到一百多Gb的存储，对于每一个节点必须保存一个区块链的完整副本。这对于大多数使用比特币的人显然不合适，于是中本聪提出了简单支付验证SPV(Simplified Payment Verification).

简单地说，SPV是一个轻量级的比特币节点，它并不需要下载区块链的所有数据内容。为了实现SPV，就需要有一个方式来检查某区块是否包含某一笔交易，这就是MerkleTree能帮我们解决的问题。

一个区块的结构里只有一个哈希值，但是这个哈希值包含了所有交易的哈希值。我们将区块内所有交易哈希的值两两进行哈希得到一个新的哈希值，然后再把得到的新的哈希值两两哈希...不断进行这个过程直到最后只存在一个哈希值。这样的结构是不是很像一颗二叉树，我们将这样的二叉树就叫做MerkleTree。

比特币的MerkleTree结构图：

![Bitcoin_MerkleTree.png](https://upload-images.jianshu.io/upload_images/830585-c44d5529ec2d62ed.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

这样，我们只需要一个根哈希就可以验证一笔交易是否存在于一个区块中了，因为这个根哈希可以遍历到所有交易哈希。相当于一个Merkle 树根哈希和一个 Merkle 路径。

# 废话少说撸代码

### MerkleTree结构

MerkleTree只包含一个根节点，每一个默克尔树节点包含数据和左右指针。每个节点都可以连接到下个节点，并依次连接到更远的节点直到叶子节点。

```
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
```

### NewMerkleNode 

新建节点时首先要从叶子节点开始创建，叶子节点只有数据，对交易哈希进行哈希得到叶子节点的数值；当创建非叶子节点时，将左右子节点的数据拼接进行哈希得到新节点的数据值。

```
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
```

### NewMerkleTree

当用叶子节点去生成一颗默克尔树时，必须保证叶子节点的数量为偶数，如果不是需要复制一份最后的交易哈西值到最后拼凑成偶数个交易哈希。

叶子节点两两哈希形成新的节点，新节点继续两两哈希直到最后只有一个节点。

```
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
```

### 集成到Block中

在Block中使用MerkleTree优化得到一个区块交易哈希的方法：

```

// 需要将Txs转换成[]byte
func (block *Block) HashTransactions() []byte  {

	//引入MerkleTree前的交易哈希
	//var txHashes [][]byte
	//var txHash [32]byte
	//
	//for _, tx := range block.Txs {
	//
	//	txHashes = append(txHashes, tx.TxHAsh)
	//}
	//txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	//
	//return txHash[:]

	//默克尔树根节点表示交易哈希
	var transactions [][]byte

	for _, tx := range block.Txs {

		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}
```

至此，大功告成。我们已将MerkleTree集成到项目中。当然，目前还看不出其作用。今天只是初窥MerkleTree的概念和用法，在之后的公链项目中再具体讲其对于区块链的妙用。






