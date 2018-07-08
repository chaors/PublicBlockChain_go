goLang公链实战之数据持久化



[TOC]
# boltDB数据库
我们知道，bitcoin客户端的区块信息是存储在LevelDB数据库中。我们既然要基于go开发公链，这里用到的数据库是基于go的[boltDB](https://github.com/boltdb)。

### 安装

使用go get
```
$ go get github.com/boltdb/boltd / ...
```

安装成功后，我们会在go目录下看到：

![boltdb安装目录](https://upload-images.jianshu.io/upload_images/830585-883a42d48e5e3518.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

### 基本使用

##### 创建并打开数据库
注意：在这里gland直接运行，生成的my.db在main.go上层目录;命令行build在运行的话是当前目录！！！
```
//1.数据库创建
 //在这里gland直接运行，生成的my.db在main.go上层目录;命令行build在运行的话是当前目录！！！
 db, err := bolt.Open("chaorsBlock.db", 0600, nil)
 if err != nil {
  log.Fatal(err)
 }
 defer db.Close()
```

在你打开之后，你有两种处理它的方式：读-写和只读操作，读-写方式开始于db.Update方法，常用于建表和表中插入新数据；只读操作开始于db.View方法，常用于表数据的查询。

##### 创建新表
```
//2.创建表
 err = db.Update(func(tx *bolt.Tx) error {
  
                //判断要创建的表是否存在
  b := tx.Bucket([]byte("MyBlocks"))
  if b == nil {
  
            //创建叫"MyBucket"的表
   _, err := tx.CreateBucket([]byte("MyBlocks"))
   if err != nil {
                                //也可以在这里对表做插入操作
    log.Fatal(err)
   }
  }

  //一定要返回nil
  return nil
 })

 //更新数据库失败
 if err != nil {
  log.Fatal(err)
 }
```

#####更新表内容
```
//3.更新表数据
 err = db.Update(func(tx *bolt.Tx) error {

  //取出叫"MyBucket"的表
  b := tx.Bucket([]byte("MyBlocks"))

  //往表里面存储数据
  if b != nil {
                        //插入的键值对数据类型必须是字节数组
   err := b.Put([]byte("l"), []byte("0x0000"))
   err := b.Put([]byte("ll"), []byte("0x0001"))
                        err := b.Put([]byte("lll"), []byte("0x0002"))
   if err != nil {
    log.Fatal(err)
   }
  }

  //一定要返回nil
  return nil
 })

 //更新数据库失败
 if err != nil {
  log.Fatal(err)
 }
```

##### 表查询
```
//4.查看表数据
 err = db.View(func(tx *bolt.Tx) error {

  //取出叫"MyBucket"的表
  b := tx.Bucket([]byte("MyBlocks"))

  //往表里面存储数据
  if b != nil {

   data := b.Get([]byte("l"))
   fmt.Printf("%s\n", data)
   data := b.Get([]byte("l"))
   fmt.Printf("%s\n", data)
  }

  //一定要返回nil
  return nil
 })

 //查询数据库失败
 if err != nil {
  log.Fatal(err)
 }
```

boltdb基本使用就先学到这，搭建公链用数据库存储区块大概也只用到这么多，以后具体涉及到boltdb其他知识再针对性学习就好。

# boltDB实现公链数据持久化。

### 存储方式

区块链的数据主要集中在各个区块上，所以区块链的数据持久化即可转化为对每一个区块的存储。boltDB是KV存储方式，因此这里我们可以以区块的哈希值为Key，区块为Value。

此外，我们还需要存储最新区块的哈希值。这样，就可以找到最新的区块，然后按照区块存储的上个区块哈希值找到上个区块，以此类推便可以找到区块链上所有的区块。

### 区块序列化

我们知道，boltDB存储的键值对的数据类型都是字节数组。所以在存储区块前需要对区块进行序列化，当然读取区块的时候就需要做反序列化处理。

没什么难点，都是借助系统方法实现。废话少说上代码。

##### 序列化
```
//区块序列化
func (block *Block) Serialize() []byte  {

 var result bytes.Buffer
 encoder := gob.NewEncoder(&result)

 err := encoder.Encode(block)
 if err != nil{

  log.Panic(err)
 }

 return result.Bytes()
}
```

##### 反序列化
```
//区块反序列化
func DeSerializeBlock(blockBytes []byte) *Block  {

 var block *Block
 dencoder := gob.NewDecoder(bytes.NewReader(blockBytes))

 err := dencoder.Decode(&block)
 if err != nil{

  log.Panic(err)
 }

 return block
}
```

### 区块链类

##### 区块链结构

之前定义的区块链结构是这样的：
```
type Blockchain struct {
 //有序区块的数组
 Blocks [] *Block
}
```
但是这样的结构，每次运行程序区块数组都是从零开始创建，并不能实现区块链的数据持久化。这里的数组属性要改为boltDB类型的区块数据库，同时还必须有一个存储当前区块链最新区块哈希的属性。
```
type Blockchain struct {
 //最新区块的Hash
 Tip []byte
 //存储区块的数据库
 DB *bolt.DB
}
```

##### 相关数据库常量
```
//相关数据库属性
const dbName = "chaorsBlockchain.db"
const blockTableName = "chaorsBlocks"
const newestBlockKey = "chNewestBlockKey"
```

##### 创建区块链
```
//1.创建带有创世区块的区块链
func CreateBlockchainWithGensisBlock() *Blockchain {

 var blockchain *Blockchain

 //判断数据库是否存在
 if IsDBExists(dbName) {

  db, err := bolt.Open(dbName, 0600, nil)
  if err != nil {
   log.Fatal(err)
  }
  err = db.View(func(tx *bolt.Tx) error {

   b := tx.Bucket([]byte(blockTableName))
   if b != nil {

    hash := b.Get([]byte(newestBlockKey))
    blockchain = &Blockchain{hash, db}
    //fmt.Printf("%x", hash)
   }

   return nil
  })
  if err != nil {

   log.Panic(err)
  }

  //blockchain.Printchain()
  //os.Exit(1)
  return blockchain
 }

 //创建并打开数据库
 db, err := bolt.Open(dbName, 0600, nil)
 if err != nil {
  log.Fatal(err)
 }
 err = db.Update(func(tx *bolt.Tx) error {

  b := tx.Bucket([]byte(blockTableName))
  //blockTableName不存在再去创建表
  if b == nil {

   b, err = tx.CreateBucket([]byte(blockTableName))
   if err != nil {

    log.Panic(err)
   }
  }

  if b != nil {

   //创世区块
   gensisBlock := CreateGenesisBlock("Gensis Block...")
   //存入数据库
   err := b.Put(gensisBlock.Hash, gensisBlock.Serialize())
   if err != nil {
    log.Panic(err)
   }

   //存储最新区块hash
   err = b.Put([]byte(newestBlockKey), gensisBlock.Hash)
   if err != nil {
    log.Panic(err)
   }

   blockchain = &Blockchain{gensisBlock.Hash, db}
  }

  return nil
 })
 //更新数据库失败
 if err != nil {
  log.Fatal(err)
 }

 return blockchain
}
```

##### 新增区块

前面我们写的这个方法为：
```
func (blc *Blockchain) AddBlockToBlockchain(data string, height int64, prevHash []byte)  {
```

仔细看发现，参数好多显得巨繁琐。那是否有些参数是没必要传递的呢？

我们既然用数据库实现了区块链的数据持久化，这里的高度height可以根据上个区块高度自增，prevHash也可以从数据库中取出上个区块而得到。因此，从今天开始，该方法省去这两个参数。

```
//2.新增一个区块到区块链
func (blc *Blockchain) AddBlockToBlockchain(data string) {

 err := blc.DB.Update(func(tx *bolt.Tx) error {

  //1.取表
  b := tx.Bucket([]byte(blockTableName))
  if b != nil {

   //2.height,prevHash都可以从数据库中取到 当前最新区块即添加后的上一个区块
   blockBytes := b.Get(blc.Tip)
   block := DeSerializeBlock(blockBytes)

   //3.创建新区快
   newBlock := NewBlock(data, block.Height+1, block.Hash)
   //4.区块序列化入库
   err := b.Put(newBlock.Hash, newBlock.Serialize())
   if err != nil {
    log.Fatal(err)
   }
   //5.更新数据库里最新区块
   err = b.Put([]byte(newestBlockKey), newBlock.Hash)
   if err != nil {
    log.Fatal(err)
   }
   //6.更新区块链最新区块
   blc.Tip = newBlock.Hash
  }

  return nil
 })
 if err != nil {
  log.Fatal(err)
 }
}
```

##### 区块链遍历

```
//3.遍历输出所有区块信息  --> 以后一般使用优化后的迭代器方法(见3.X)
func (blc *Blockchain) Printchain() {

 var block *Block
 //当前遍历的区块hash
 var curHash []byte = blc.Tip
 for {

  err := blc.DB.View(func(tx *bolt.Tx) error {

   b := tx.Bucket([]byte(blockTableName))
   if b != nil {
    blockBytes := b.Get(curHash)
    block = DeSerializeBlock(blockBytes)

    /**时间戳格式化 Format里的年份必须是固定的！！！
    这个好像是go诞生的时间
    time.Unix(block.Timestamp, 0).Format("2006-01-02 15:04:05")
    "2006-01-02 15:04:05"格式固定，改变其他也可能会出错
    */
    fmt.Printf("\n#####\nHeight:%d\nPrevHash:%x\nHash:%x\nData:%s\nTime:%s\nNonce:%d\n#####\n",
     block.Height, block.PrevBlockHash, block.Hash, block.Data, time.Unix(block.Timestamp, 0).Format("2006-01-02 15:04:05"), block.Nonce)
   }
   return nil
  })
  if err != nil {
   log.Fatal(err)
  }

  var hashInt big.Int
  hashInt.SetBytes(block.PrevBlockHash)
  //遍历到创世区块，跳出循环  创世区块哈希为0
  if big.NewInt(0).Cmp(&hashInt) == 0 {

   break
  }
  curHash = block.PrevBlockHash
 }
}
```

###### 注意：
###### time.Unix(block.Timestamp, 0).Format("2006-01-02 15:04:05")  goLang这里真是奇葩啊……时间戳格式化只能写"2006-01-02 15:04:05"，一个数丢不能写错，不然你会”被穿越“的！！！据说这个日期是go语言的诞生日期，还真是傲娇啊，生怕大家不知道吗？？？

##### 判断区块链数据库是否存在
```
//判断数据库是否存在
func IsDBExists(dbName string) bool {

 //if _, err := os.Stat(dbName); os.IsNotExist(err) {
 //
 // return false
 //}

 _, err := os.Stat(dbName)
 if err == nil {

  return true
 }
 if os.IsNotExist(err) {

  return false
 }

 return true
}
```


### 区块链迭代器

对区块链区块的遍历上面已经实现，但是还可以优化。我们不难发现区块链的区块遍历类似于单向链表的遍历，那么我们能不能制造一个像链表的Next属性似的迭代器，只要通过不断地访问Next就能遍历所有的区块？

话都说到这份上了，答案当然是肯当的。

#####BlockchainIterator
```
//区块链迭代器
type BlockchainIterator struct {
 //当前遍历hash
 CurrHash []byte
 //区块链数据库
 DB *bolt.DB
}
```

##### Next迭代方法
```

func (blcIterator *BlockchainIterator) Next() *Block {

 var block *Block
 //数据库查询
 err := blcIterator.DB.View(func(tx *bolt.Tx) error {

  b := tx.Bucket([]byte(blockTableName))
  if b != nil {

   //获取当前迭代器对应的区块
   currBlockBytes := b.Get(blcIterator.CurrHash)
   block = DeSerializeBlock(currBlockBytes)

   //更新迭代器
   blcIterator.CurrHash = block.PrevBlockHash
  }
  return nil
 })
 if err != nil {
  log.Fatal(err)
 }

 return block
}
```

##### 怎么用？

###### 1.在Blockchain类新增一个生成当前区块链的迭代器的方法
```
//生成当前区块链迭代器的方法
func (blc *Blockchain) Iterator() *BlockchainIterator {

 return &BlockchainIterator{blc.Tip, blc.DB}
}
```

###### 2.修改之前的Printchain方法
```
//3.X 优化区块链遍历方法
func (blc *Blockchain) Printchain() {
 //迭代器
 blcIterator := blc.Iterator()
 for {

  block := blcIterator.Next()
  fmt.Printf("\n#####\nHeight:%d\nPrevHash:%x\nHash:%x\nData:%s\nTime:%s\nNonce:%d\n#####\n",
   block.Height, block.PrevBlockHash, block.Hash, block.Data, time.Unix(block.Timestamp, 0).Format("2006-01-02 15:04:05"),block.Nonce)

  var hashInt big.Int
  hashInt.SetBytes(block.PrevBlockHash)

  if big.NewInt(0).Cmp(&hashInt) == 0 {

   break
  }
 }
}
```
是不是发现遍历区块的代码相对简洁了，这里把数据库访问和区块迭代的代码分离到了BlockchainIterator里实现，也符合程序设计的单一职责原则。

###main函数测试
```

package main

import (
 "chaors.com/LearnGo/publicChaorsChain/part4-DataPersistence-Prototype/BLC"
)

func main() {

 blockchain := BLC.CreateBlockchainWithGensisBlock()
 defer blockchain.DB.Close()

 //添加一个新区快
 blockchain.AddBlockToBlockchain("first Block")
 blockchain.AddBlockToBlockchain("second Block")
 blockchain.AddBlockToBlockchain("third Block")

 blockchain.Printchain()

}
```
##### 1.首次运行(这时不存在数据库)

![mainTest_1](https://upload-images.jianshu.io/upload_images/830585-bb2ba4ba0fcf0c66.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

##### 2.注释掉三句AddBlockToBlockchain代码，再次运行

![mainTest_2](https://upload-images.jianshu.io/upload_images/830585-332ef0509d7cc2be.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

这次我们并没有添加区块，所以打印区没有挖矿的过程。但是打印的区块是上次AddBlockToBlockchain添加的，说明区块存储成功了。

##### 2.修改AddBlockToBlockchain段代码，再次运行
```
blockchain.AddBlockToBlockchain("4th Block")
blockchain.AddBlockToBlockchain("5th Block")
```

![mainTest_3](https://upload-images.jianshu.io/upload_images/830585-1f14259a42c23d25.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

我们看到，在原有区块信息不变的情况，新挖出的区块成功添加到了区块链数据库中。说明我们的区块链数据持久化实现成功了。

