# goLang公链实战之交易(2)



前面最开始构建了交易的基本模型，然后逐步实现了转账，集成了钱包地址。公链基本交易模块已然成型，还有些小的细节需要去实现和优化。

### Coinbase奖励

我们知道，在比特币中每当矿工成功挖到一个区块，就会得到一笔奖励。这笔奖励包含在创币交易中，创币交易是一个区块中的第一笔交易。

目前的项目还没有引入多节点竞争挖矿，暂且认为每一个区块是转账的第一个发起人挖到的。如此，Coinbase奖励就应该添加到MineNewBlock方法里。

```
//2.新增一个区块到区块链 --> 包含交易的挖矿
func (blc *Blockchain) MineNewBlock(from []string, to []string, amount []string) {

	//send -from '["chaors"]' -to '["xyx"]' -amount '["5"]'

	//获取UTXO集
	utxoSet := &UTXOSet{blc}

	var txs []*Transaction

	//作为奖励给矿工的奖励  暂时将这笔奖励给from[0]  挖矿成功后再转给挖矿的矿工
	tx := NewCoinbaseTransaction(from[0])
	txs = append(txs, tx)

	//1.通过相关算法建立Transaction数组
	for index, address := range from {

		value, _ := strconv.Atoi(amount[index])
		tx := NewTransaction(address, to[index], int64(value), utxoSet, txs)
		txs = append(txs, tx)
	}
	...
	...
}
```

# UTXOSet

之前为了实现转账引入了UTXO的概念，但是我们每次在转账查询可用余额时，都会去遍历一遍数据库上的区块。这样，会随着区块链的不断扩张，转账时查询的成本越来越高。

毕竟，查询时我们只需要关注未花费的TxOutput信息，而不需要关注区块上其他信息。那么，我们为什么不把未花费的TxOutput信息单独存储来查询呢？

其实，比特币正是这样做的。Bitcoin将链上所有区块存储在blocks数据库,将所有UTXO的集存储在chainstate数据库。

于是，我们引入UTXOSet集用来实现UTXO集的数据库存储。

```
//存储未花费交易输出的数据库表
const UTXOTableName  = "UTXOTableName"

type UTXOSet struct {

	Blockchain *Blockchain
}
```

## ResetUTXOSet

该方法初始化UTXO集的存储表，如果bucket 存在就先移除，然后从区块链中获取所有的未花费输出，最终将输出保存到 bucket 中。

```
// 重置数据库表
func (utxoSet *UTXOSet) ResetUTXOSet()  {

	err := utxoSet.Blockchain.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(utxoTableName))

		if b != nil {


			err := tx.DeleteBucket([]byte(utxoTableName))

			if err!= nil {
				log.Panic(err)
			}

		}

		b ,_ = tx.CreateBucket([]byte(utxoTableName))
		if b != nil {

			//[string]*TXOutputs
			txOutputsMap := utxoSet.Blockchain.FindUTXOMap()


			for keyHash,outs := range txOutputsMap {

				txHash,_ := hex.DecodeString(keyHash)

				b.Put(txHash,outs.Serialize())

			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}
```

既然是UTXO存储表的初始化，那么一般情况下它只被执行一次。这种特性和创世区块相似，所以我们需要把他的实现放到创世区块的创建中。

```
//1.创建创世区块
func CreateBlockchainWithGensisBlock(address string) *Blockchain {

	...
	...
	
	//创建创世区块时候初始化UTXO表
	utxoSet := &UTXOSet{blc}
	utxoSet.ResetUTXOSet()

	return blc
}
```

## UTXO表存储格式

我们存储UTXO的目的也是为了转账时能够更快地查询余额，为了实现转账，UTXO表除了存储对应的未花费的TXOutput集，还需要存储这些TXOutput来自于哪一笔交易。

我们以交易的哈希为键，以该交易下TXOutput组成的数组为值来存储UTXO。

由于一个交易下可能存在多个TXOutput，显然可以用数组表示。我们引入TXOutputs类来表示，因为要存储这些TXOutput需要实现序列化。

#### TXOutputs
```
type TXOutputs struct {

	UTXOS []*UTXO
}


// 序列化成字节数组
func (txOutputs *TXOutputs) Serialize() []byte {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(txOutputs)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// 反序列化
func DeserializeTXOutputs(txOutputsBytes []byte) *TXOutputs {

	var txOutputs TXOutputs

	decoder := gob.NewDecoder(bytes.NewReader(txOutputsBytes))
	err := decoder.Decode(&txOutputs)
	if err != nil {

		log.Panic(err)
	}

	return &txOutputs
}
```
#### FindUTXOMap
知道UTXO表的存储格式后，我们就需要能够找到区块链上所有的未花费的TXOutput，并且能够以UTXO表存储的格式返回。这个方法是基于Blockchain的。

```
// 查找未花费的UTXO[string]*TXOutputs 返回字典  键为所属交易的哈希，值为TXOutput数组
func (blc *Blockchain) FindUTXOMap() map[string]*TXOutputs  {

	blcIterator := blc.Iterator()

	// 存储已花费的UTXO的信息
	spentableUTXOsMap := make(map[string][]*TXInput)

	utxoMaps := make(map[string]*TXOutputs)


	for {

		block := blcIterator.Next()

		for i := len(block.Txs) - 1; i >= 0 ;i-- {

			txOutputs := &TXOutputs{[]*UTXO{}}
			tx := block.Txs[i]

			// coinbase
			if tx.IsCoinbaseTransaction() == false {

				for _,txInput := range tx.Vins {

					txHash := hex.EncodeToString(txInput.TxHash)
					spentableUTXOsMap[txHash] = append(spentableUTXOsMap[txHash],txInput)
				}
			}

			txHash := hex.EncodeToString(tx.TxHAsh)

		WorkOutLoop:
			for index,out := range tx.Vouts  {

				txInputs := spentableUTXOsMap[txHash]

				if len(txInputs) > 0 {

					isUnSpent := true

					for _,in := range  txInputs {

						outPublicKey := out.Ripemd160Hash
						inPublicKey := in.PublicKey

						if bytes.Compare(outPublicKey,Ripemd160Hash(inPublicKey)) == 0{

							if index == in.Vout {

								isUnSpent = false
								continue WorkOutLoop
							}
						}

					}

					if isUnSpent {

						utxo := &UTXO{tx.TxHAsh,index,out}
						txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
					}

				} else {

					utxo := &UTXO{tx.TxHAsh,index,out}
					txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
				}

			}

			// 设置键值对
			utxoMaps[txHash] = txOutputs
		}


		// 找到创世区块时退出
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {

			break
		}
	}

	return utxoMaps
}
```
#### UTXOSet测试

我们在命令行新增一个命令用于打印目前UTXO表存储的内容。

test命令添加和解析

```
testCmd := flag.NewFlagSet("test", flag.ExitOnError)

...
...

switch os.Args[1] {
     ...
     ...
	case "test":
		//第二个参数为相应命令，取第三个参数开始作为参数并解析
		err := testCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
}

...
...

//UTXOSet测试
if testCmd.Parsed() {

	cli.TestMethod()
}
```

CLI_test.go

```
func (cli *CLI) TestMethod()  {


	fmt.Println("TestMethod")

	blockchain := GetBlockchain()
	defer blockchain.DB.Close()

	utxoSet := &UTXOSet{blockchain}
	utxoSet.ResetUTXOSet()

	fmt.Println(blockchain.FindUTXOMap())
}
```



## GetBalance

现在我们查询余额就不需要去遍历整个区块数据库了，只需要遍历存储UTXO的表即可。查询逻辑还是和之前一样，先要从表中查到对应地址的所有UTXO，然后累加他们的值。

```
// 3.查询余额
func (utxoSet *UTXOSet) GetBalance(address string) int64 {

	UTXOS := utxoSet.FindUTXOsForAddress(address)

	var amount int64

	for _, utxo := range UTXOS  {

		amount += utxo.Output.Value
	}

	return amount
}

// 2.查询某个地址的UTXO
func (utxoSet *UTXOSet) FindUTXOsForAddress(address string) []*UTXO {

	var utxos []*UTXO

	err := utxoSet.Blockchain.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(UTXOTableName))

		// 游标
		c := b.Cursor()
		for k, v := c.First(); k != nil; k,v = c.Next() {

			txOutputs := DeserializeTXOutputs(v)

			for _, utxo := range txOutputs.UTXOS {

				if utxo.Output.UnLockScriptPubKeyWithAddress(address) {

					utxos = append(utxos,utxo)
				}
			}
		}

		return nil
	})
	if err != nil {

		log.Panic(err)
	}

	return utxos
}
```

接下来，修改CLI中查询余额调用的方法即可。

```

//查询余额
func (cli *CLI) getBlance(address string) {

	fmt.Println("地址：" + address)

	blockchain := GetBlockchain()
	defer blockchain.DB.Close()

	//amount := blockchain.GetBalance(address)  引入UTXOSet前的方法

	utxoSet := &UTXOSet{blockchain}
	amount := utxoSet.GetBalance(address)

	fmt.Printf("%s一共有%d个Token\n", address, amount)
}
```

## send转账优化

我们来回忆一下转账的几个重要步骤：
> 1.找到发起方所有的UTXO
> 2.在1的UTXO集里找到符合该次转账条件的UTXO组合和该组合对应的代币总和。
> 3.发起转账

之前上面的1，2都是在Blockchain里实现，并需要遍历整个区块数据库。在引入UTXOSet之后就需要基于UTXOSet实现。

> 1.Blockchain.UTXOs --> UTXOSet.FindUnPackageSpendableUTXOS

> 2.Blockchain.FindSpendableUTXOs --> UTXOSet.FindSpendableUTXOs

显然，2的方法基本相同。但是1有所差别，因为之前的逻辑中Blockchain.UTXOs需要找到链上所有UTXO和未打包的UTXO，而引入UTXOSet之后链上的UTXO都在UTXO表中，我们只需要找到未打包的交易产生的UTXO即可。

找到未打包交易的UTXO
```
// 找到未打包交易的UTXO，对应TXOutput的TX的Hash和index
func (utxoSet *UTXOSet) FindUnPackageSpendableUTXOS(address string, txs []*Transaction) []*UTXO {

	var unUTXOs []*UTXO
	spentTXOutputs := make(map[string][]int)

	for _,tx := range txs {

		if tx.IsCoinbaseTransaction() == false {

			for _, in := range tx.Vins {

				//是否能够解锁
				if in.UnlockWithAddress(address) {

					key := hex.EncodeToString(in.TxHash)
					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
				}
			}
		}
	}

	for _,tx := range txs {

	Work:
		for index,out := range tx.Vouts {

			if out.UnLockScriptPubKeyWithAddress(address) {

				if len(spentTXOutputs) != 0 {

					for hash,indexArray := range spentTXOutputs {

						txHashStr := hex.EncodeToString(tx.TxHAsh)

						if hash == txHashStr {

							var isUnSpent =true
							for _,outIndex := range indexArray {

								if index == outIndex {

									isUnSpent = false
									continue Work
								}

								if isUnSpent {

									utxo := &UTXO{tx.TxHAsh, index, out}
									unUTXOs = append(unUTXOs, utxo)
								}
							}
						} else {

							utxo := &UTXO{tx.TxHAsh, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				} else {

					utxo := &UTXO{tx.TxHAsh, index, out}
					unUTXOs = append(unUTXOs, utxo)
				}
			}
		}
	}

	return unUTXOs
}
```

找到未花费交易里满足当次交易的UTXO组合

```
//转账时查找可用的用于消费的UTXO组合
func (utxoSet *UTXOSet) FindSpendableUTXOs(address string,amount int64,txs []*Transaction) (int64,map[string][]int)  {

	unPackageUTXOS := utxoSet.FindUnPackageSpendableUTXOS(address, txs)

	spentableUTXO := make(map[string][]int)

	var value int64 = 0

	for _, UTXO := range unPackageUTXOS {

		value += UTXO.Output.Value
		txHash := hex.EncodeToString(UTXO.TxHash)
		spentableUTXO[txHash] = append(spentableUTXO[txHash], UTXO.Index)

		if value >= amount{

			return  value, spentableUTXO
		}
	}

	// 钱还不够
	err := utxoSet.Blockchain.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(UTXOTableName))

		if b != nil {

			c := b.Cursor()
		UTXOBREAK:
			for k, v := c.First(); k != nil; k, v = c.Next() {

				txOutputs := DeserializeTXOutputs(v)

				for _, utxo := range txOutputs.UTXOS {

					value += utxo.Output.Value
					txHash := hex.EncodeToString(utxo.TxHash)
					spentableUTXO[txHash] = append(spentableUTXO[txHash], utxo.Index)

					if value >= amount {

						break UTXOBREAK
					}
				}
			}
		}

		return nil
	})
	if err != nil {

		log.Panic(err)
	}

	if value < amount{

		fmt.Printf("%s found.余额不足...", value)
		os.Exit(1)
	}

	return  value, spentableUTXO
}
```

## UTXOSet.Update

由于转账消耗了一定的UTXO，同时产生了一定的UTXO。所以转账之后需要对UTXO数据库表做更新以保持UTXO表存储的永远是最新的未花费的交易输出。

简单地梳理一下更新UTXO表的步骤:

> 1.找到最新添加到区块链上的区块
> 
> 2.遍历区块交易，将所有交易输入集中到一个数组
> 
> 3.遍历区块交易的交易输出，找到新增的未花费的TXOutput
> 
> 4.在UTXO表中删除输入输入中已花费的TXOutput，并将未花费的TXOutput缓存
> 
> 5.将3求出的和4缓存的TXOutput新增到UTXO表中

废话少说撸代码

```
//更新UTXO 
func (utxoSet *UTXOSet) Update()  {

	// 1.找出最新区块
	block := utxoSet.Blockchain.Iterator().Next()

	// 未花费的UTXO  键为对应交易哈希，值为TXOutput数组
	outsMap := make(map[string] *TXOutputs)
	// 新区快的交易输入,这些交易输入引用的TXOutput被消耗，应该从UTXOSet删除
	ins := []*TXInput{}

	// 2.遍历区块交易找出交易输入
	for _, tx := range block.Txs {

		//遍历交易输入，
		for _, in := range tx.Vins {

			ins = append(ins, in)
		}
	}

	// 2.遍历交易输出
	for _, tx := range block.Txs {

		utxos := []*UTXO{}

		for index, out := range tx.Vouts {

			//未花费标志
			isUnSpent := true
			for _, in := range ins {

				if in.Vout == index && bytes.Compare(tx.TxHAsh, in.TxHash) == 0 &&
					bytes.Compare(out.Ripemd160Hash, Ripemd160Hash(in.PublicKey)) == 0 {

						isUnSpent = false
						continue
				}
			}

			if isUnSpent {

				utxo := &UTXO{tx.TxHAsh,index,out}
				utxos = append(utxos,utxo)
			}
		}

		if len(utxos) > 0 {

			txHash := hex.EncodeToString(tx.TxHAsh)
			outsMap[txHash] = &TXOutputs{utxos}
		}
	}

	//3. 删除已消耗的TXOutput
	err := utxoSet.Blockchain.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(UTXOTableName))
		if b != nil {

			for _, in := range ins {

				txOutputsBytes := b.Get(in.TxHash)

				//如果该交易输入无引用的交易哈希
				if len(txOutputsBytes) == 0 {

					continue
				}
				txOutputs := DeserializeTXOutputs(txOutputsBytes)

				// 判断是否需要
				isNeedDelete := false

				//缓存来自该交易还未花费的UTXO
				utxos := []*UTXO{}

				for _, utxo := range txOutputs.UTXOS {

					if in.Vout == utxo.Index && bytes.Compare(utxo.Output.Ripemd160Hash, Ripemd160Hash(in.PublicKey)) == 0 {

						isNeedDelete = true
					}else {

						//txOutputs中剩余未花费的txOutput
						utxos = append(utxos,utxo)
					}
				}

				if isNeedDelete {

					b.Delete(in.TxHash)

					if len(utxos) > 0 {

						preTXOutputs := outsMap[hex.EncodeToString(in.TxHash)]
						preTXOutputs.UTXOS = append(preTXOutputs.UTXOS, utxos...)
						outsMap[hex.EncodeToString(in.TxHash)] = preTXOutputs
					}
				}
			}

			// 4.新增交易输出到UTXOSet
			for keyHash, outPuts := range outsMap {

				keyHashBytes, _ := hex.DecodeString(keyHash)
				b.Put(keyHashBytes, outPuts.Serialize())
			}
		}

		return nil
	})
	if err != nil{

		log.Panic(err)
	}
}
```

### Main_Test

```
package main

import (

	"chaors.com/LearnGo/publicChaorsChain/part11-transaction_1_Prototype/BLC"
)

func main() {

	cli := BLC.CLI{}
	cli.Run()
}
```

创建区块链

![main_test1.png](https://upload-images.jianshu.io/upload_images/830585-b9c0fecbbb8a5a2c.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

产生交易后

![main_test2.png](https://upload-images.jianshu.io/upload_images/830585-eca5facd400159f4.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)


到目前为止，公链的交易算是基本讲完了。




