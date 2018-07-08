# goLang公链实战之转账(2)

上节已基本实现硬编码转账并查询余额，今天真正地实现转账函数并对相关函数做一个优化。

### UTXO
UTXO 代表 Unspent Transaction TxOutput,表示区块链上未经花费的交易输出。简单地说，UTXO还没有被包含在任何的交易输入中。根据UTXO可以知道对应TxOutput来自哪一笔交易，以及其在Vounts中的下标。

```
type UTXO struct {
	//来自交易的哈希
	TxHash []byte
	//在该交易VOuts里的下标
	Index int
	//未花费的交易输出
	Output *TXOutput
}
```
### UTXOs函数改造

有了UTXO的结构后，我们就可以改造上次获取未花费输出的方法，使其返回为UTXO类型的数组。

其次，之前测试的都是单笔转账的交易。当出现多笔转账的交易时，我们现有的查询余额方法会不准确。为什么呢？

当一笔交易中有多个转账，当进行其中第二笔转账时，第一笔转账已经成功。但是，我们此时查询的依然是区块链上所有交易的UTXO。因此，我们还需要在UTXOs方法中加上当前未上链的所有交易的UTXO。

这时就有疑问了，不是只有上链的交易才会有效吗？事实是这样的，但是看目前的项目，由于还没有引入竞争挖矿的概念，每一次send必然会挖矿成功，其交易必然会上链。所以我们需要暂时这么做。

```
//5.返回一个地址对应的UTXO的交易UTXOs
//func (blc *Blockchain) UnSpentTransactionsWithAddress(address string) []*Transaction {
func (blc *Blockchain) UTXOs(address string, txs []*Transaction) []*UTXO {

	//未花费的TXOutput
	var utxos []*UTXO

	//已经花费的TXOutput [hash:[]] [交易哈希：TxOutput对应的index]
	var spentTXOutputs = make(map[string][]int)

	//遍历器处理区块链上的UTXO
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
		Work:
			for index, out := range tx.Vouts {

				//验证当前输出是否是
				if out.UnLockScriptPubKeyWithAddress(address) {

					//fmt.Println(out)
					//fmt.Println(spentTXOutputs)

					//判断是否曾发生过交易
					if spentTXOutputs != nil {

						if len(spentTXOutputs) != 0 {

							//未花费UTXO标志
							isUnSpentUTXO := true

							//遍历spentTXOutputs
							for txHash, indexArray := range spentTXOutputs {

								//遍历TXOutputs下标数组
								for _, i := range indexArray {

									if index == i && txHash == hex.EncodeToString(tx.TxHAsh) {

										isUnSpentUTXO = false
										continue Work
									}
								}
							}

							if isUnSpentUTXO {

								utxo := &UTXO{tx.TxHAsh, index, out}
								utxos = append(utxos, utxo)
							}
						} else {

							utxo := &UTXO{tx.TxHAsh, index, out}
							utxos = append(utxos, utxo)
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

	//处理未打包到区块链上的交易集里的UTXO
	for _, tx := range txs {

		if tx.IsCoinbaseTransaction() == false {
			for _, in := range tx.Vins {

				if in.UnlockWithAddress(address) {

					key := hex.EncodeToString(in.TxHash)

					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
				}
			}
		}
	}

	for _, tx := range txs {
	Work1:
		for index, out := range tx.Vouts {

			if out.UnLockScriptPubKeyWithAddress(address) {

				if len(spentTXOutputs) != 0 {

					for hash, indexArray := range spentTXOutputs {

						txHashStr := hex.EncodeToString(tx.TxHAsh)

						if hash == txHashStr {

							isUnSpentUTXO := true

							for _, outIndex := range indexArray {

								if index == outIndex {

									isUnSpentUTXO = false
									continue Work1
								}

								if isUnSpentUTXO {

									utxo := &UTXO{tx.TxHAsh, index, out}
									utxos = append(utxos, utxo)
								}
							}
						} else {

							utxo := &UTXO{tx.TxHAsh, index, out}
							utxos = append(utxos, utxo)
						}
					}
				} else {

					utxo := &UTXO{tx.TxHAsh, index, out}
					utxos = append(utxos, utxo)
				}
			}
		}
	}

	return utxos
}
```

### TXInput和TXOutput解锁

上面UTXOs方法求得是某一个address的所有UTXO，目前我们还没有引入钱包地址的概念，姑且理解这个address为用户名。我们要想保证查询的是某个用户(address)交易输入和输出是属于这个用户的，必须有一个保障的机制。

```

//验证当前输入是否是当前地址的
func (txInput *TXInput) UnlockWithAddress(address string) bool  {

	return txInput.ScriptSig == address
}

//验证当前交易输出属于某用户
func (txOutput *TXOutput) UnLockScriptPubKeyWithAddress(address string) bool {

	return txOutput.ScriptPubKey == address
}
```
### FindSpendableUTXOs

当我们进行一笔转账时，交易输入有可能引用一个UTXO，也可能引用多个UTXO。在获取转账方所有的UTXO后，还需要找到符合条件的UTXO组合作为交易输入的引用。这个时候可能出现用户余额不足以转账的情况，也可能出现UTXO组合价值大于转账金额产生找零的情况。

为了方便地判断UTXO来源以及计算转账后的找零，我们需要想办法在当前用户的所有UTXO中找到一个满足当前转账情况的UTXO集，并返回其UTXO总额和对应的UTXO集。而这个UTXO集是一个字典类型，键是UTXO来源交易的哈希，值对该交易下UTXO对应TXOutput在Vounts中的下标。

```
//转账时查找可用的用于消费的UTXO  返回输入总金额和一个字典，UTXO集是一个字典类型，键是UTXO来源交易的哈希，值对该交易下UTXO对应TXOutput在Vounts中的下标
func (blc *Blockchain) FindSpendableUTXOs(address string, amount int, txs []*Transaction) (int64, map[string][]int) {

	//1.获取当前地址所有UTXO
	utxos := blc.UTXOs(address, txs)
	spendableUTXO := make(map[string][]int)

	//2.遍历UTXO
	//总的金额
	var value int64
	for _, utxo := range utxos {

		value += utxo.Output.Value
		txHash := hex.EncodeToString(utxo.TxHash)
		spendableUTXO[txHash] = append(spendableUTXO[txHash], utxo.Index)

		if value >= int64(amount) {

			break
		}
	}

	//余额不足
	if value < int64(amount) {

		fmt.Println("%s found.余额不足...", value)
		os.Exit(1)
	}

	return value, spendableUTXO
}
```
### NewTransaction

上次我们硬编码测试了几笔交易，这回有了上面的基础方法就可以对普通交易的构造做一个代码实现。

```
//2.普通交易
func NewTransaction(from string, to string, amount int, blc *Blockchain, txs []*Transaction) *Transaction {

	//获取from用户用于这笔交易的总输入金额和UTXO集
	money, spendableUTXODic := blc.FindSpendableUTXOs(from, amount, txs)

	//输入输出
	var txInputs []*TXInput
	var txOutputs []*TXOutput

	//遍历spendableUTXODic来组装TXInput作为该交易的交易输入
	for txHash, indexArr := range spendableUTXODic {

		//字符串转换为[]byte
		txHashBytes, _ := hex.DecodeString(txHash)
		for _, index := range indexArr {

			//交易输入
			txInput := &TXInput{
				txHashBytes,
				index,
				from,
			}
			txInputs = append(txInputs, txInput)
		}
	}

	//转账
	txOutput := &TXOutput{
		int64(amount),
		to,
	}
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = &TXOutput{
		money-int64(amount),
		from,
	}
	txOutputs = append(txOutputs, txOutput)

	//交易构造
	tx := &Transaction{
		[]byte{},
		txInputs,
		txOutputs,
	}

	tx.HashTransactions()

	return tx
}
```

### MineNewBlock

理论上我们的交易是支持多笔转账的，可是上面构建交易的方法是针对一笔交易。所以，我们需要在发起交易挖掘区块的方法里对cli输入的多笔交易信息做一个遍历并生成多笔交易数据。

```
//2.新增一个区块到区块链 --> 包含交易的挖矿
func (blc *Blockchain) MineNewBlock(from []string, to []string, amount []string) {

	//send -from '["chaors"]' -to '["xyx"]' -amount '["5"]'

	//1.通过相关算法建立Transaction数组
	var txs []*Transaction

	//遍历输入输出，组装多笔交易
	for index, address := range from {

		value, _ := strconv.Atoi(amount[index])
		tx := NewTransaction(address, to[index], value, blc, txs)
		txs = append(txs, tx)
	}

	//2.挖矿
	//取上个区块的哈希和高度值
	var block *Block
	err := blc.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {

			hash := b.Get([]byte(newestBlockKey))
			blockBytes := b.Get(hash)
			block = DeSerializeBlock(blockBytes)
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

			//fmt.Printf("444---%x\n\n", block.Txs[0].Vins[0].TxHash)
			//fmt.Println(block)

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

### CLI优化

上面已经基本实现了多笔交易的打包并挖矿。接下来，我们看一下CLI.go文件的结构：

```
type CLI struct {

}

//打印目前左右命令使用方法
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tcreateBlockchain -address --创世区块地址 ")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT --交易明细")
	fmt.Println("\tprintchain --打印所有区块信息")
	fmt.Println("\tgetbalance -address -- 输出区块信息.")
}

func isValidArgs() {

	//获取当前输入参数个数
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {

	isValidArgs()

	//自定义cli命令
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

不难返现逻辑不是很清晰，既有cli命令的定义和解析，又有具体命令的实现。按照单一职责的设计原则，这里应该只有cli命令的定义和解析，具体命令的解析应该拆分到相应文件。这样显得脉络清晰，逻辑明了。

例如，我们可以吧创建区块链命令的具体实现分离到一个CLI_createBlockchain.go文件：
```
//新建区块链
func (cli *CLI)creatBlockchain(address string)  {

	blockchain := CreateBlockchainWithGensisBlock(address)
	defer blockchain.DB.Close()
}
```
目前CLI支持的命令还有，打印区块链(printchain)，获取余额(getBlance)，转账(send)，我们按照上面的处理方式分别把代码分离就可以了。最终，项目会多出这几个文件：

![CLI优化.png](https://upload-images.jianshu.io/upload_images/830585-6f7679637acc5517.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

至此，就基本实现了公链的转账功能。








