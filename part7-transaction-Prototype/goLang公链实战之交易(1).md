# goLang公链实战之交易(1)


区块链区块的作用是打包链上产生的交易,可以说交易是区块链至关重要的一个组成部分.在区块链中,交易一旦被创建，就没有任何人能够再去修改或是删除它.

关于交易的结构,可以参照以前的这篇:[比特币源码研读(3)数据结构-交易Transaction](https://www.jianshu.com/p/3447aab7e864)

我们之前定义的区块结构简单地用一个字符串描述交易内容,今后将正式用新的结构来表示交易:
```
type Block struct {
	//1.区块高度
	Height int64
	//2.上一个区块HAsh
	PrevBlockHash []byte
	//3.交易数据
	//Data []byte    //只前简单地对交易的描述
        Txs [] *Transaction   //Transaction结构数组表示交易
	//4.时间戳
	Timestamp int64
	//5.Hash
	Hash []byte
	//6.Nonce  符合工作量证明的随机数
	Nonce int64
}
```
### Transaction结构
```
type Transaction struct {
	//1.交易哈希值
	TxHAsh []byte
	//2.交易输入
	Vins []*TXInput
	//3.交易输出
	Vouts []*TXOutput
}
```
### TXInput
```
type TXInput struct {
	//交易ID 引用上一笔交易输出作为输入
	TxHash []byte
	//存储TXOutput在Vout里的索引
	Vout int
	//数字签名  暂时可理解为用户名
	ScriptSig string
}
```
交易输入作为本次交易的消费源,输入来源于之前交易的输出.如上,TxHash是引用的上一笔输出所在的交易的交易哈希;Vout是该输出在相应交易中的输出索引;ScriptSig,可以暂时理解为用户名,表示哪一个用户拥有这一笔输入.ScriptSig的设定是为了保证用户只能话费自己名下的代币.

### TXOutput
```
type TXOutput struct {
	//面值
	Value int64
	//暂时理解为用户名
	ScriptPubKey string
}
```
这里的交易输出就是上面交易输入里引用的输出.Value是该输出的面值,ScriptPubKey暂时理解为用户名,表示谁将拥有这笔输出.

了解比特币的人都知道,交易输出是一个完整的不可分割的结构.什么意思呢?就是我们在引用TXOutput,必须全部引用,不能仅仅使用其一部分.举个简单的🌰:

假如你有一个25btc的TXOutput,你需要花费10btc.这个交易的过程并不是:你花费了25btc中的10btc,你的原有TXOutput依旧有15btc的余额.真正的过程是,你花费了整个原有的TXOutput,由于消费额不匹配,这里会产生一个15btc的找零.消费的结果是:你25btc的TXOutput被话费已不复存在,系统重新为你生成一个15btc面值的TXOutput.这两个TXOutput是完全不同的两个对象!!!

### CoinbaseTransaction

我们知道,当矿工成功挖到一个区块时会获得一笔奖励.那么这笔奖励是怎么交付到矿工账户的.这就有赖于一笔叫做创币交易的交易.

创币交易是区块内的第一笔交易,它负责将系统产生的奖励给挖出区块的矿工.由于它并不是普通意义上的转账,所以交易输入里并不需要引用任何一笔交易输出.

```
//创建创币交易
func NewCoinbaseTransaction(address string) *Transaction {

	//交易输入  由于区块第一笔交易其实没有输入，所以交易哈希传空，TXOutput索引传-1，签名随你
	txInput := &TXInput{
		[]byte{},
		-1,
		"CoinbaseTransaction",
	}

	//交易输出  产生一笔奖励给挖矿者address
	txOutput := &TXOutput{25, address}
	txCoinbase := &Transaction{
		[]byte{},	//暂时将交易哈希置空
		[]*TXInput{txInput},
		[]*TXOutput{txOutput},
	}
		
	//交易哈希的计算
	txCoinbase.HashTransactions()

	return txCoinbase
}
```

### HashTransactions
```
//将交易信息转换为字节数组
func (tx *Transaction) HashTransactions() {

	//交易信息序列化
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
	if err != nil {

		log.Panic(err)
	}

	//设置hash
	txHash := sha256.Sum256(result.Bytes())
	tx.TxHAsh = txHash[:]
}
```
### NewBlock改动
```
//1.创建新的区块
func NewBlock(txs []*Transaction, height int64, prevBlockHash []byte) *Block {

	//创建区块
	block := &Block{
		Height:        height,
		PrevBlockHash: prevBlockHash,
		Txs:           txs,
		Timestamp:     time.Now().Unix(),
		Hash:          nil,
		Nonce:         0}

	//调用工作量证明返回有效的Hash
	pow := NewProofOfWork(block)
	hash, nonce := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	fmt.Printf("\r######%d-%x\n", nonce, hash)

	return block
}
```

### CreateBlockchainWithGensisBlock改动
```
//1.创建创世区块
func CreateBlockchainWithGensisBlock(address string) {

	//判断数据库是否存在
	if IsDBExists(dbName) {

		fmt.Println("创世区块已存在...")
		os.Exit(1)

		//创建并打开数据库
		db, err := bolt.Open(dbName, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}

		var block *Block
		err = db.View(func(tx *bolt.Tx) error {

			b := tx.Bucket([]byte(blockTableName))
			if b != nil {

				hash := b.Get([]byte(newestBlockKey))
				block = DeSerializeBlock(b.Get(hash))
				fmt.Printf("\r######%d-%x\n", block.Nonce, hash)
			}

			return nil
		})
		if err != nil {

			log.Panic(err)
		}

		os.Exit(1)
	}

	fmt.Println("正在创建创世区块...")

	//创建并打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucket([]byte(blockTableName))
		if err != nil {

			log.Panic(err)
		}

		if b != nil {

			//创币交易
			txCoinbase := NewCoinbaseTransaction(address)
			//创世区块
			gensisBlock := CreateGenesisBlock([]*Transaction{txCoinbase})
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
		}

		return nil
	})
	//更新数据库失败
	if err != nil {
		log.Fatal(err)
	}
}
```
### POW/prepareData改动

添加交易后,POW挖矿时也必须相应地把交易信息考虑进去.这里需要改动prepareData方法
```
//拼接区块属性，返回字节数组
func (pow *ProofOfWork) prepareData(nonce int) []byte {

	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,
			pow.Block.HashTransactions(),
			IntToHex(pow.Block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
			IntToHex(int64(pow.Block.Height)),
		},
		[]byte{},
	)

	return data
}
```
这里POW计算目标哈希并不需要将所有的交易信息拼接,我们只需要将每一个交易的交易哈希拼接起来即可.因为,交易哈希是交易所有信息的哈希值.这样做也能保证交易信息的完整性.所以,我们需要在Block新增一个方法:

```
//将交易信息转换为字节数组
func (block *Block) HashTransactions() []byte  {

	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range block.Txs {

		txHashes = append(txHashes, tx.TxHAsh)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}
```

### Printchain添加交易信息打印

现在,整个交易的数据结构就搭起来了.我们再改动区块链打印方法,将区块的交易信息添加到打印.

```
//3.X 优化区块链遍历方法
func (blc *Blockchain) Printchain() {
	//迭代器
	blcIterator := blc.Iterator()
	for {

		block := blcIterator.Next()

		fmt.Println("------------------------------")
		fmt.Printf("Height：%d\n", block.Height)
		fmt.Printf("PrevBlockHash：%x\n", block.PrevBlockHash)
		fmt.Printf("Timestamp：%s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash：%x\n", block.Hash)
		fmt.Printf("Nonce：%d\n", block.Nonce)
		fmt.Println("Txs:")
		for _,tx := range block.Txs {

			fmt.Printf("%x\n", tx.TxHAsh)
			fmt.Println("Vins:")
			for _,in := range tx.Vins  {
				fmt.Printf("txHash:%x\n", in.TxHash)
				fmt.Printf("Vout:%d\n", in.Vout)
				fmt.Printf("ScriptSig:%s\n\n", in.ScriptSig)
			}

			fmt.Println("Vouts:")
			for _,out := range tx.Vouts  {
				fmt.Printf("Value:%x\n", out.Value)
				fmt.Printf("ScriptPubKey:%x\n\n", out.ScriptPubKey)
			}
		}
		fmt.Println("------------------------------")

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if big.NewInt(0).Cmp(&hashInt) == 0 {

			break
		}
	}
}
```
### Main_Test
```
package main

import (
	"chaors.com/LearnGo/publicChaorsChain/part7-transaction-Prototype/BLC"
)

func main() {

	cli := BLC.CLI{}
	cli.Run()
}
```
打印创世区块的结果为:

![main_test](https://upload-images.jianshu.io/upload_images/830585-286f383c9d56f375.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)







