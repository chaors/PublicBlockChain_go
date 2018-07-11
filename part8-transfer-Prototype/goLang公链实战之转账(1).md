# goLang公链实战之转账(1)
上次基本实现了交易的数据结构,在此基础上便可以来实现转账,即区块链的普通交易.

### cli转账命令

我们知道挖矿的目的是找到一个公认的记账人把当前的所有交易打包到区块并添加到区块链上.之前我们使用addBlock命令实现添加区块到区块链的,这里转账包含挖矿并添加到区块链.所以,我们需要在cli工具类里用转账命令send代替addBlock命令.

其次我们都知道,一次区块可以包括多个交易.因此,这里我们的转账命令要设计成支持多笔转账.


```
//命令说明方法 打印目前左右命令使用方法
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tcreateBlockchain -address --创世区块地址 ")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT --交易明细")
	fmt.Println("\tprintchain --打印所有区块信息")
	fmt.Println("\tgetbalance -address -- 输出区块信息.")
}
```
```
func (cli *CLI) Run() {

	isValidArgs()

	//自定义cli命令
	//转账
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printchainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createBlockchain", flag.ExitOnError)
	blanceBlockCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)

	//addBlockCmd 设置默认参数
	flagSendBlockFrom := sendBlockCmd.String("from", "", "源地址")
	flagSendBlockTo := sendBlockCmd.String("to", "", "目标地址")
	flagSendBlockAmount := sendBlockCmd.String("amount", "", "转账金额")
	flagCreateBlockchainAddress := createBlockchainCmd.String("address", "", "创世区块地址")
	flagBlanceBlockAddress := blanceBlockCmd.String("address", "", "输出区块信息")

	//解析输入的第二个参数是addBlock还是printchain，第一个参数为./main
	switch os.Args[1] {
	case "send":
		//第二个参数为相应命令，取第三个参数开始作为参数并解析
		err := sendBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createBlockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getBalance":
		err := blanceBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1)
	}

	//对addBlockCmd命令的解析
	if sendBlockCmd.Parsed() {

		if *flagSendBlockFrom == "" {

			printUsage()
			os.Exit(1)
		}
		if *flagSendBlockTo == "" {

			printUsage()
			os.Exit(1)
		}
		if *flagSendBlockAmount == "" {

			printUsage()
			os.Exit(1)
		}

		//cli.addBlock(*flagAddBlockData)

		//这里真正地调用转账方法
		//fmt.Println(*flagSendBlockFrom)
		//fmt.Println(*flagSendBlockTo)
		//fmt.Println(*flagSendBlockAmount)
		//
		//fmt.Println(Json2Array(*flagSendBlockFrom))
		//fmt.Println(Json2Array(*flagSendBlockTo))
		//fmt.Println(Json2Array(*flagSendBlockAmount))
		cli.send(
			Json2Array(*flagSendBlockFrom),
			Json2Array(*flagSendBlockTo),
			Json2Array(*flagSendBlockAmount),
			)
	}
	//对printchainCmd命令的解析
	if printchainCmd.Parsed() {

		cli.printchain()
	}
	//
	if createBlockchainCmd.Parsed() {

		if *flagCreateBlockchainAddress == "" {

			cli.creatBlockchain(*flagCreateBlockchainAddress)
		}

		cli.creatBlockchain(*flagCreateBlockchainAddress)
	}

	if blanceBlockCmd.Parsed() {

		if *flagBlanceBlockAddress == "" {

			printUsage()
			os.Exit(1)
		}

		cli.getBlance(*flagBlanceBlockAddress)
	}
}
```
### Json2Arr

命令行输入的都是字符串,要想让转账命令支持多笔转账,则输入的信息是json形式的数组.在编码实现解析并转账的时候,我们需要将Json字符串转化为数组类型.这个功能在utils里实现.

我们一般输入的转账命令是这样的:
```
send -from '["chaors", "ww"]' -to '["xyz", "dh"]' -amount '["5", "100"]'
```
>send       转账命令
from       发送方
to           接收方
amount  转账金额
三个参数的数组分别一一对应,上述命令表示:
chaors转给xyx共5btc;
ww转给dh的100btc.

utils.go
```
// 标准的JSON字符串转数组
func Json2Array(jsonString string) []string {

	//json 到 []string
	var sArr []string
	if err := json.Unmarshal([]byte(jsonString), &sArr); err != nil {

		log.Panic(err)
	}
	return sArr
}
```

### 转账的理解

说到转账,就离不开交易.这里的转账便是普通交易,之前我们只实现了创币交易.这里需要实现普通交易.

为了更好地理解转账的过程,我们先将复杂问题简单化.假设每一个区块只有一笔交易,我们看一个简单的小🌰.

1.节点chaors挖到一个区块，产生25BTC的创币交易。由于是创币交易，其本身是不需要引用任何交易输出的，所以在输入对象TXInput的交易哈希为空，vount所在的下标为-1，数字签名为空或者随便填写；输出对象里btc拥有者为chaors，面值为25btc  创世区块交易结构
```
 txInput0 = &TXInput{[]byte{},-1,"Gensis Block"}
 txOutput0 = &TXOutput{25, "chaors"}  //在gaVouts索引为0

 CoinbaseTransaction{"00000",
			[]*TXInput{txInput0},
			[]*TXOutput{txOutput0}
}
```

2.chaors获得25btc后，他的好友ww知道后向他索要10btc.大方的chaors便把10btc转给ww.此时
 交易的输入为chaors上笔交易获得的btc,TXInput对象的交易ID为奖励chaors的上一个交易ID，vount下标为chaors的TXOutput下标，签名此时且认为是来自chaors，填作"chaors" 此时chaors的25btc面值的TXOutput就被花费不复存在了，那么chaors还应该有15btc的找零哪去了？系统会为chaors的找零新生成一个面值15btc的TXOutput。所以，这次有一个输入，两个输出。
> chaors(25) 给 ww 转 10 -- >>  chaors(15) + ww(10)

这次的交易结构为:
```
 //输入
 txInput1 = &TXInput{"00000",0,"chaors"}
 //"00000" 相当于来自于哈希为"00000"的交易
 //索引为零，相当于上一次的txOutput0为输入

 //输出
 txOutput1 = &TXOutput{10, "ww"}		//在该笔交易Vouts索引为0  chaors转给ww的10btc产生的输出
 txOutput2 = &TXOutput{15, "chaors"}    //在该笔交易Vouts索引为1  给ww转账产生的找零
 Transaction1{"11111"，
			[]*TXInput{txInput1}
			[]*TXOutput{txOutput1, txOutput2}
}
```

3.ww感觉拥有比特币是一件很酷的事情，又来跟chaors要。出于兄弟情谊，chaors又转给ww7btc
这次的交易结构为:
```
//输入
 txInput2 = &TXInput{"11111",2,"chaors"}

 //输出
 txOutput3 = &TXOutput{7, "ww"}		  //在该笔交易Vouts索引为0
 txOutput4 = &TXOutput{8, "chaors"}   //在该笔交易Vouts索引为1
 Transaction2{"22222"，
			[]*TXInput{txInput2}
			[]*TXOutput{txOutput3, txOutput4}
}
```

4.消息传到他们共同的朋友xyz那里，xyz觉得btc很好玩向ww索要15btc.ww一向害怕xyx，于是尽管不愿意也只能屈服。

我们来看看ww此时的所有财产:
```
txOutput1 = &TXOutput{10, "ww"}		//来自Transaction1(hash:11111)Vouts索引为0的输出   
txOutput3 = &TXOutput{7, "ww"}		//来自Transaction2(hash:2222)Vouts索引为0的输出
```
想要转账15btc,ww的哪一笔txOutput都不够，这个时候就需要用ww的两个txOutput都作为
 输入,这次的交易结构为:
```
//输入：
txInput3 = &TXInput{"11111",1,"ww"}
txInput4 = &TXInput{"22222",3,"ww"}

//输出
 txOutput5 = &TXOutput{15, "xyz"}		索引为5
 txOutput6 = &TXOutput{2, "ww"}        索引为6

 第四个区块交易结构
 Transaction3{"33333"，
			[]*TXInput{txInput3, txInput4}
			[]*TXOutput{txOutput5, txOutput6}
}
```
现在,我们来总结一下上述几个交易.
> A.chaors
>> 1.从CoinbaseTransaction获得TXOutput0总额25
>> 2.Transaction1转给ww10btc,TXOutput0被消耗,获得txOutput2找零15btc
>> 3.Transaction2转给ww7Btc,txOutput2被消耗,获得txOutput4找零8btc
>> 4.最后只剩8btc的txOutput4作为未花费输出

> B.ww
>> 1.从Transaction1获得TXOutput1,总额10btc
>> 2.从Transaction2获得TXOutput3,总额7btc
>> 3.Transaction3转给xyz15btc,TXOutput1和TXOutput3都被消耗,获得txOutput6找零2btc
>> 4.最后只剩2btc的txOutput6作为未花费输出

> C.xyz
>> 1.从Transaction3获得TXOutput5,总额15btc
>> 2.拥有15btc的TXOutput5作为未花费输出

经过这个例子,我们可以发现转账具备几个特点:
##### 1.每笔转账必须有输入TXInput和输出TXOutput
##### 2.每笔输入必须有源可查(TXInput.TxHash)
##### 3.每笔输入的输出引用必须是未花费的(没有被之前的交易输入所引用)
##### 4.TXOutput是一个不可分割的整体,一旦被消耗就不可用.消费额度不对等时会有找零(产生新的TXOutput)

这个🌰很重要,对于后面转账的代码逻辑是个扎实的基础准备.

### AddBlockToBlockchain --> MineNewBlock

既然在cli工具用转账命令send代替了添加区块,那么在实际的函数调用中,我们必须考虑到交易信息.上面对转账有了一定的理解,现在可以认为构造公链的第一笔交易.

```
//2.普通交易
func NewTransaction(from []string, to []string, amount []string) *Transaction {


	//单笔交易构造假数据测试交易

	//输入输出
	var txInputs []*TXInput
	var txOutputs []*TXOutput

	//输入,由于这里引用的TXOutput来自创世区块的奖励, 这里复制创世区块里创币交易的哈希作为交易输入对TXOutput的引用
	txHash, _ := hex.DecodeString("d3c17e00ad2c1bd7fec8f5afde710f2c3afd40478c3cca492d7e9a2b0cbe4808")
	txInput := &TXInput {
		txHash,
		0, 	//要花费的TXOutput在对应交易的Vounts下标为0
		from[0],
	}

	fmt.Printf("111--%x\n", txInput.TxHash)

	txInputs = append(txInputs, txInput)

	//转账
	txOutput := &TXOutput{
		10,
	to[0],
	}
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = &TXOutput{
		25-10,
		from[0],
	}
	txOutputs = append(txOutputs, txOutput)

	tx := &Transaction{
		[]byte{},
		txInputs,
		txOutputs,
	}

	tx.HashTransactions()

	fmt.Printf("222---%x\n", txInput.TxHash)

	return tx


	//1. 有一个函数，返回from这个人所有的未花费交易输出所对应的Transaction
	//unSpentTx := UnSpentTransactionsWithAddress("chaors")
	//fmt.Println(unSpentTx)
}
```
我们在人为通过硬编码构造好一个基于创世区块的转账交易后,此时需要将这笔交易打包到区块并添加到区块链上.之前我们的AddBlockToBlockchain就需要做些改动.

```
//2.新增一个区块到区块链 --> 包含交易的挖矿
//func (blc *Blockchain) AddBlockToBlockchain(txs []*Transaction) {
func (blc *Blockchain) MineNewBlock(from []string, to []string, amount []string) {

	//send -from '["chaors"]' -to '["xyx"]' -amount '["5"]'

	tx := NewTransaction(from, to, amount)
	//1.通过相关算法建立Transaction数组
	var txs []*Transaction
	txs = append(txs, tx)

	fmt.Printf("333---%x\n\n", txs[0].Vins[0].TxHash)

	//2.挖矿
	//取上个区块的哈希和高度值
	var block *Block
	err := blc.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			hash := b.Get([]byte(newestBlockKey))
			block = DeSerializeBlock(b.Get(hash))
		}

		return nil
	})
	if err != nil {

		log.Panic(err)
	}

	//3.建立新区块
	block = NewBlock(txs, block.Height+1, block.Hash)

	//4.存储新区块
	err = blc.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			fmt.Printf("444---%x\n\n", block.Txs[0].Vins[0].TxHash)
			fmt.Println(block)

			err = b.Put(block.Hash, block.Serialize())
			if err != nil {

				log.Panic(err)
			}

			err = b.Put([]byte(newestBlockKey), block.Hash)
			if err != nil {

				log.Panic(err)
			}

			blc.Tip = block.Hash
		}

		return nil
	})
	if err != nil {

		log.Panic(err)
		//fmt.Print(err)
	}
}
```
然后再CLI工具将send命令的具体实现添加好.
```
//转账
func (cli *CLI) send(from []string, to []string, amount []string)  {

	blockchain := GetBlockchain()
	defer blockchain.DB.Close()

	blockchain.MineNewBlock(from, to, amount)
}
```

### 余额查询GetBalance

转账和区块上链都实现了,Run也是没有问题的.那么怎么验证转账成功呢?毕竟此时我们不知道各自的余额是多少.这时候,我们就需要来实现余额查询方法.

首先在CLi工具里实现,getBlance命令的添加和解析其实在前面说到send命令时已经有了.回去看看即可.

余额查询的实现
```
//余额查询
func (cli *CLI) getBlance(address string) {

	fmt.Println("地址：" + address)

	blockchain := GetBlockchain()
	defer blockchain.DB.Close()

	amount := blockchain.GetBalance(address)

	fmt.Printf("%s一共有%d个Token\n", address, amount)
}
```
```

//查询余额
func (blc *Blockchain) GetBalance(address string) int64 {

	utxos := blc.UTXOs(address)

	var amount int64
	for _, out := range utxos {

		amount += out.Value
	}

	return amount
}
```

### UTXOs

要想实现余额查询,必须知道某个账户未花费的TxOutput.这个时候我们需要遍历区块链上的区块,然后去每一笔交易里找.在每笔交易输出里被引用的TxOutput必定被消耗,只需要记录被消耗的TxOutput.然后再去比对每笔交易产生的TxOutput,做个去除即可得到当前账户在链上剩余的未花费的TxOutput.

```
//5.返回一个地址对应的UTXO的交易UTXOs
//func (blc *Blockchain) UnSpentTransactionsWithAddress(address string) []*Transaction {
func (blc *Blockchain) UTXOs(address string) []*TXOutput {

	//未花费的TXOutput
	var UTXOs []*TXOutput

	//已经花费的TXOutput [hash:[]] [交易哈希：TxOutput对应的index]
	var spentTXOutputs = make(map[string][]int)

	//遍历器
	blcIterator := blc.Iterator()

	for {

		block := blcIterator.Next()

		//fmt.Println(block)
		//fmt.Println()

		for _, tx := range block.Txs {

			// txHash

			// Vins
			//判断当前交易是否为创币交易
			if tx.IsCoinbaseTransaction() == false {

				for _, in := range tx.Vins {

					//验证当前输入是否是当前地址的
					if in.UnlockWithAddress(address) {

						key := hex.EncodeToString(in.TxHash)

						//fmt.Printf("lll%x\n", in.TxHash)
						//fmt.Println(key)
						spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
					}

				}
			}


			// Vouts
			for index, out := range tx.Vouts {

				//验证当前输出是否是
				if out.UnLockScriptPubKeyWithAddress(address) {

					fmt.Printf("%x", block.Hash)
					fmt.Println(index, out)

					//判断是否曾发生过交易
					if spentTXOutputs != nil {

						if len(spentTXOutputs) != 0 {

							//遍历spentTXOutputs
							for txHash, indexArray := range spentTXOutputs {

								//遍历TXOutputs下标数组
								for _, i := range indexArray {

									fmt.Printf("%d--%d\n", index, i)
									fmt.Printf("%s\n", txHash)
									fmt.Printf("%x\n", tx.TxHAsh)
									fmt.Println(spentTXOutputs)
									fmt.Println(out)

									if index == i && txHash == hex.EncodeToString(tx.TxHAsh) {

										continue
									} else {

										//fmt.Println(index,i)
										//fmt.Println(out)
										//fmt.Println(spentTXOutputs)

										UTXOs = append(UTXOs, out)
									}
								}
							}
						} else {

							UTXOs = append(UTXOs, out)
						}
					}
				}
			}
		}

		//找到创世区块，跳出循环
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		// Cmp compares x and y and returns:
		//
		//   -1 if x <  y
		//    0 if x == y
		//   +1 if x >  y
		if hashInt.Cmp(big.NewInt(0)) == 0 {

			break
		}
	}

	return UTXOs
}
```
### Main_test
```
package main

import (

	"chaors.com/publicChaorsChain/part8-transfer-Prototype/BLC"
)

func main() {

	cli := BLC.CLI{}
	cli.Run()

	//blc := BLC.CreateBlockchainWithGensisBlock("chaors")
	//utxos := blc.UnUTXOs("chaors")
	//fmt.Println(utxos)
}
```
创建好创世区块后,执行第一次转账.
chaors(25btc) -->ww(10) + chaors(15)

![Main_test1.png](https://upload-images.jianshu.io/upload_images/830585-8db744d967a117f8.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

![Main_test2.png](https://upload-images.jianshu.io/upload_images/830585-6d8e26a9ca70d86e.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

### 总结

当前的交易函数是人工硬编码，下次再具体实现。


















	

