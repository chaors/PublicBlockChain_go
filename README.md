# PublicBlockChain_go 
# 基于goLang的公链实现(具备公链全功能)

# 前言

之前在github上看到一位歪果友人Ivan Kuznetsov用go实现的公链Demo，正好自己也在学习区块链相关知识，于是想记录自己是如何一步步实现基于GoLang的公链项目。喜欢的朋友可以给个star权当鼓励，也可以一起探讨区块链相关知识，我们共同进步。

# 目录

### 1.[区块链基础结构(part1)](https://github.com/chaors/PublicBlockChain_go/blob/master/part1-Basic-Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8B%E5%8C%BA%E5%9D%97%E9%93%BE%E5%9F%BA%E7%A1%80%E7%BB%93%E6%9E%84.md)

### 2.[ProofOfWork(part2)](https://github.com/chaors/PublicBlockChain_go/blob/master/part2-ProofOfWork-Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8BProofOfWork.md)

### 3.[BoltDB(part3)](https://github.com/chaors/PublicBlockChain_go/blob/master/part3-boltdb-Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8Bboltdb%E6%95%B0%E6%8D%AE%E5%BA%93.md)

### 4.[链上数据持久化(part4)](https://github.com/chaors/PublicBlockChain_go/blob/master/part4-DataPersistence-Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8B%E6%95%B0%E6%8D%AE%E6%8C%81%E4%B9%85%E5%8C%96.md)

### 5.[CLI命令行工具(part5/part6)](https://github.com/chaors/PublicBlockChain_go/blob/master/part5-cli-Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8Bcli%E5%B7%A5%E5%85%B7.md)

### 6.[交易TransAction1(part7)](https://github.com/chaors/PublicBlockChain_go/blob/master/part7-transaction-Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8B%E4%BA%A4%E6%98%93(1).md)

### 7.[转账1(part8)](https://github.com/chaors/PublicBlockChain_go/blob/master/part8-transfer-Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8B%E8%BD%AC%E8%B4%A6(1).md)

### 8.[转账2(part9)](https://github.com/chaors/PublicBlockChain_go/blob/master/part9-transfer_1-Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8B%E8%BD%AC%E8%B4%A6(2).md)

### 9.[钱包地址Address(part10)](https://github.com/chaors/PublicBlockChain_go/blob/master/part10-address_Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8B%E9%92%B1%E5%8C%85%26%E5%9C%B0%E5%9D%80.md)

### 10.[交易TransAction2(part11)](https://github.com/chaors/PublicBlockChain_go/blob/master/part11-transaction_1_Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8B%E4%BA%A4%E6%98%93(2).md)

### 11.[MerkleTree(part12)](https://github.com/chaors/PublicBlockChain_go/blob/master/part12-MerkleTree_Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8BMerkleTree.md)

### 12.还在陆续更新中....

### ......



# 更多区块链技术文章请访问[chaors](https://www.jianshu.com/c/6277257ba30a)


# 参考资料

### 1. [Building Blockchain in Go-Ivan Kuznetso](https://jeiwan.cc/tags/blockchain/)

### 2.[用 golang 从零开始构建区块链(Bitcoin)系列](https://liuchengxu.gitbooks.io/blockchain-tutorial/content/)

### 3.[《精通比特币第二版》](http://book.8btc.com/books/6/masterbitcoin2cn/_book/trans-preface.html)

### 4.[Bitcoin Developer Documentation](https://bitcoin.org/en/developer-documentation)

### 5.[bitcoin wiki](https://en.bitcoin.it/wiki/Main_Page)

