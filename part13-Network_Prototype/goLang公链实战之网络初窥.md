# 

前面实现了公链的基本结构，交易，钱包地址，数据持久化，交易等功能。但显然这些功能都是基于单节点的，我们都知道比特币网络是一个多节点共存的P2P网络。

比特币网络上的节点主要有以下几类(图片来自《精通比特币》)：

![比特币网络节点.png](https://upload-images.jianshu.io/upload_images/830585-d0677abe2a0dcadf.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

> M:矿工节点，具备挖矿功能的节点。这些节点一般运行在特殊的硬件设备以完成复杂的工作量证明运算。有些矿工节点同时也是全节点。

> W:钱包节点，常见的很多比特币客户端属于钱包节点，它不需要拷贝完整的区块链。一般的钱包节点都是SPV节点，SPV节点借助之前讲的MerkleTree原理使得不需要下载所有区块就能验证交易成为可能，后面讲到钱包开发再深入理解。

>  B:全节点具有完整的，最新的区块链拷贝。可以独立自主地校验所有交易。

## 复杂问题简单化

由于P2P网络的复杂性，为了便于理解区块链网络同步的原理，我们可以将复杂的网络简单化为只有三个核心节点的网络：

***1.中心节点(全节点)：其他节点会连接到这个节点来更新区块数据***
***2.钱包节点：用于钱包之间实现交易，但这里它依旧存储一个区块链的完整副本***
***3.矿工节点：矿工节点会在内存池中存储交易并在适当时机将交易打包挖出一个新区块 但这里它依旧存储一个区块链的完整副本***

我们在这个简化基础上去实现区块链的网络同步。

# 几个重要的数据结构

要想实现数据的同步，必须有两个节点间的通讯。那么他们通讯的内容和格式是什么样的呢？

区块链同步时两个节点的通讯信息并不是单一的，不同的情况和不同的阶段通讯的格式与处理方式是不同的。这里分析主要用的几个数据结构。

为了区分节点发送的信息，我们需要定义几个消息类型来区别他们。

```
package BLC

// 采用TCP
const PROTOCOL  = "tcp"
// 发送消息的前12个字节指定了命令名(version)
const COMMANDLENGTH  = 12
// 节点的区块链版本
const NODE_VERSION  = 1

// 命令
// 版本命令
const COMMAND_VERSION  = "version"
const COMMAND_ADDR  = "addr"
const COMMAND_BLOCK  = "block"
const COMMAND_INV  = "inv"
const COMMAND_GETBLOCKS  = "getblocks"
const COMMAND_GETDATA  = "getdata"
const COMMAND_TX  = "tx"

// 类型
const BLOCK_TYPE  = "block"
const TX_TYPE  = "tx"
```



## Version

Version消息是发起区块同步第一个发送的消息类型，其内容主要有区块链版本，区块链最大高度，来自的节点地址。它主要用于比较两个节点间谁是最长链。

```
type Version struct {
	// 区块链版本
	Version    int64
	// 请求节点区块的高度
	BestHeight int64
	// 请求节点的地址
	AddrFrom   string
}
```

组装发送Version信息

```
//发送COMMAND_VERSION
func sendVersion(toAddress string, blc *Blockchain)  {


	bestHeight := blc.GetBestHeight()
	payload := gobEncode(Version{NODE_VERSION, bestHeight, nodeAddress})

	request := append(commandToBytes(COMMAND_VERSION), payload...)

	sendData(toAddress, request)
}
```

当一个节点收到Version信息，会比较自己的最大区块高度和请求者的最大区块高度。如果自身高度大于请求节点会向请求节点回复一个版本信息告诉请求节点自己的相关信息；否则直接向请求节点发送一个GetBlocks信息。

```
// Version命令处理器
func handleVersion(request []byte, blc *Blockchain)  {

	var buff bytes.Buffer
	var payload Version

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	// 提取最大区块高度作比较
	bestHeight := blc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if bestHeight > foreignerBestHeight {

		// 向请求节点回复自身Version信息
		sendVersion(payload.AddrFrom, blc)
	} else if bestHeight < foreignerBestHeight {

		// 向请求节点要信息
		sendGetBlocks(payload.AddrFrom)
	}

// 添加到已知节点中
	if !nodeIsKnown(payload.AddrFrom) {

		knowedNodes = append(knowedNodes, payload.AddrFrom)
	}
}
```

Blockchain获取自身最大区块高度的方法：

```
// 获取区块链最大高度
func (blc *Blockchain) GetBestHeight() int64 {

	block := blc.Iterator().Next()

	return block.Height
}
```

## GetBlocks

当一个节点知道对方节点区块链最新，就需要发送一个GetBlocks请求来请求对方节点所有的区块哈希。这里有人觉得为什么不直接返回对方节点所有新区块呢，可是万一两个节点区块数据相差很大，在一次请求中发送相当大的数据肯定会使通讯出问题。

```
// 表示向节点请求一个块哈希的表，该请求会返回所有块的哈希
type GetBlocks struct {
	//请求节点地址
	AddrFrom string
}
```

组装发送GetBlocks消息

```
//发送COMMAND_GETBLOCKS
func sendGetBlocks(toAddress string)  {

	payload := gobEncode(GetBlocks{nodeAddress})

	request := append(commandToBytes(COMMAND_GETBLOCKS), payload...)

	sendData(toAddress, request)
}
```
当一个节点收到一个GetBlocks消息，会将自身区块链所有区块哈希算出并组装在Inv消息中发送给请求节点。一般收到GetBlocks消息的节点为较新区块链。

```
func handleGetblocks(request []byte, blc *Blockchain)  {

	var buff bytes.Buffer
	var payload GetBlocks

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := blc.GetBlockHashes()

	sendInv(payload.AddrFrom, BLOCK_TYPE, blocks)
}
```

Blockchain获得所有区块哈希的方法：

```
// 获取区块所有哈希
func (blc *Blockchain) GetBlockHashes() [][]byte {

	blockIterator := blc.Iterator()

	var blockHashs [][]byte

	for {

		block := blockIterator.Next()
		blockHashs = append(blockHashs, block.Hash)

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {

			break
		}
	}

	return blockHashs
}
```

## Inv消息

Inv消息用于收到GetBlocks消息的节点向其他节点展示自己拥有的区块或交易信息。其主要结构包括自己的节点地址，展示信息的类型，是区块还是交易，当用于节点请求区块同步时是区块信息；当用于节点向矿工节点转发交易时是交易信息。

```
// 向其他节点展示自己拥有的区块和交易
type Inv struct {
	// 自己的地址
	AddrFrom string
	// 类型 block tx
	Type     string
	// hash二维数组
	Items    [][]byte
}
```

组装发送Inv消息：

```
//COMMAND_Inv
func sendInv(toAddress string, kind string, hashes [][]byte) {

	payload := gobEncode(Inv{nodeAddress,kind,hashes})

	request := append(commandToBytes(COMMAND_INV), payload...)

	sendData(toAddress, request)
}
```

当一个节点收到Inv消息后，会对Inv消息的类型做判断分别采取处理。
如果是Block类型，它会取出最新的区块哈希并组装到一个GetData消息返回给来源节点，这个消息才是真正向来源节点请求新区块的消息。

由于这里将源节点(比当前节点拥有更新区块链的节点)所有区块的哈希都知道了，所以需要每处理一次Inv消息后将剩余的区块哈希缓存到unslovedHashes数组，当unslovedHashes长度为零表示处理完毕。

这里可能有人会有疑问，我们更新的应该是源节点拥有的新区块(自身节点没有)，这里为啥请求的是全部呢？这里的逻辑是这样的，请求的时候是请求的全部，后面在真正更新自身数据库的时候判断是否为新区块并保存到数据库。其实，我们都知道两个节点的区块最大高度，这里也可以完全请求源节点的所有新区块哈希。为了简单，这里先暂且这样处理。

如果收到的Inv是交易类型，取出交易哈希，如果该交易不存在于交易缓冲池，添加到交易缓冲池。这里的交易类型Inv一般用于有矿工节点参与的通讯。因为在网络中，只有矿工节点才需要去处理交易。

```
func handleInv(request []byte, blc *Blockchain)  {

	var buff bytes.Buffer
	var payload Inv

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// Ivn 3000 block hashes [][]
	if payload.Type == BLOCK_TYPE {

		fmt.Println(payload.Items)

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, BLOCK_TYPE , blockHash)

		if len(payload.Items) >= 1 {

			unslovedHashes = payload.Items[1:]
		}
	}

	if payload.Type == TX_TYPE {

		txHash := payload.Items[0]

		// 添加到交易池
		if mempool[hex.EncodeToString(txHash)].TxHAsh == nil {

			sendGetData(payload.AddrFrom, TX_TYPE, txHash)
		}
	}
}
```

## GetData消息

GetData消息是用于真正请求一个区块或交易的消息类型，其主要结构为：

```
// 用于请求区块或交易
type GetData struct {
	// 节点地址
	AddrFrom string
	// 请求类型  是block还是tx
	Type     string
	// 区块哈希或交易哈希
	Hash       []byte
}
```

组装并发送GetData消息。

```
func sendGetData(toAddress string, kind string ,blockHash []byte) {

	payload := gobEncode(GetData{nodeAddress,kind,blockHash})

	request := append(commandToBytes(COMMAND_GETDATA), payload...)

	sendData(toAddress, request)
}
```

当一个节点收到GetData消息，如果是请求区块，节点会根据区块哈希取出对应的区块封装到BlockData消息中发送给请求节点；如果是请求交易，同理会根据交易哈希取出对应交易封装到TxData消息中发送给请求节点。

```
func handleGetData(request []byte, blc *Blockchain)  {

	var buff bytes.Buffer
	var payload GetData

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	if payload.Type == BLOCK_TYPE {

		block, err := blc.GetBlock([]byte(payload.Hash))
		if err != nil {

			return
		}

		sendBlock(payload.AddrFrom, block)
	}

	if payload.Type == TX_TYPE {

		// 取出交易
		txHash := hex.EncodeToString(payload.Hash)
		tx := mempool[txHash]

		sendTx(payload.AddrFrom, &tx)
	}
}
```

Blockchain的GetBlock方法：

```
// 获取对应哈希的区块
func (blc *Blockchain) GetBlock(bHash []byte) ([]byte, error)  {

	//blcIterator := blc.Iterator()
	//var block *Block = nil
	//var err error = nil
	//
	//for {
	//
	//	block = blcIterator.Next()
	//	if bytes.Compare(block.Hash, bHash) == 0 {
	//
	//		break
	//	}
	//}
	//
	//if block == nil {
	//
	//	err = errors.New("Block is not found")
	//}
	//
	//return block, err

	var blockBytes []byte

	err := blc.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))

		if b != nil {

			blockBytes = b.Get(bHash)
		}

		return nil
	})

	return blockBytes, err
}
```

## BlockData
BlockData消息用于一个节点向其他节点发送一个区块，到这里才真正完成区块的发送。

```
// 用于节点间发送一个区块
type BlockData struct {
	// 节点地址
	AddrFrom string
	// 序列化区块
	BlockBytes []byte
}
```
BlockData的发送：
```
func sendBlock(toAddress string, blockBytes []byte)  {


	payload := gobEncode(BlockData{nodeAddress,blockBytes})

	request := append(commandToBytes(COMMAND_BLOCK), payload...)

	sendData(toAddress, request)
}
```
当一个节点收到一个Block信息，它会首先判断是否拥有该Block，如果数据库没有就将其添加到数据库中(AddBlock方法)。然后会判断unslovedHashes(之前缓存所有主节点未发送的区块哈希数组)数组的长度，如果数组长度不为零表示还有未发送处理的区块，节点继续发送GetData消息去请求下一个区块。否则，区块同步完成，重置UTXO数据库。

```
func handleBlock(request []byte, blc *Blockchain)  {

	//fmt.Println("handleblock:\n")
	//blc.Printchain()

	var buff bytes.Buffer
	var payload BlockData

	dataBytes := request[COMMANDLENGTH:]

	// 反序列化
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	block := DeSerializeBlock(payload.BlockBytes)
	if block == nil {

		fmt.Printf("Block nil")
	}

	err = blc.AddBlock(block)
	if err != nil {

		log.Panic(err)
	}
	fmt.Printf("add block %x succ.\n", block.Hash)
	//blc.Printchain()

	if len(unslovedHashes) > 0 {

		sendGetData(payload.AddrFrom, BLOCK_TYPE, unslovedHashes[0])
		unslovedHashes = unslovedHashes[1:]
	}else {

		//blc.Printchain()

		utxoSet := &UTXOSet{blc}
		utxoSet.ResetUTXOSet()
	}
}
```

## TxData消息

TxData消息用于真正地发送一笔交易。当对方节点发送的GetData消息为Tx类型，相应地会回复TxData消息。

```
// 同步中传递的交易类型
type TxData struct {
	// 节点地址
	AddFrom string
	// 交易
	TransactionBytes []byte
}
```

TxData消息的发送：

```
func sendTx(toAddress string, tx *Transaction)  {

	data := TxData{nodeAddress, tx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes(COMMAND_TX), payload...)

	sendData(toAddress, request)
}
```

当一个节点收到TxData消息，这个节点一般为矿工节点，如果不是他会以Inv消息格式继续转发该交易信息到矿工节点。矿工节点收到交易，当交易池满足一定数量时开始打包挖矿。

当生成新的区块并打包到区块链上时，矿工节点需要以BlockData消息向其他节点转发该新区块。

```
func handleTx(request []byte, blc *Blockchain)  {

	var buff bytes.Buffer
	var payload TxData

	dataBytes := request[COMMANDLENGTH:]
	buff.Write(dataBytes)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {

		log.Panic(err)
	}

	tx := DeserializeTransaction(payload.TransactionBytes)
	memTxPool[hex.EncodeToString(tx.TxHAsh)] = tx

	// 自身为主节点，需要将交易转发给矿工节点
	if nodeAddress == knowedNodes[0] {

		for _, node := range knowedNodes {

			if node != nodeAddress && node != payload.AddFrom {

				sendInv(node, TX_TYPE, [][]byte{tx.TxHAsh})
			}
		}
	} else {

		//fmt.Println(len(memTxPool), len(miningAddress))
		if len(memTxPool) >= minMinerTxCount && len(miningAddress) > 0 {

		MineTransactions:

			var txs []*Transaction
			// 创币交易，作为挖矿奖励
			coinbaseTx := NewCoinbaseTransaction(miningAddress)
			txs = append(txs, coinbaseTx)

			var _txs []*Transaction

			for id := range memTxPool {

				tx := memTxPool[id]
				_txs = append(_txs, &tx)
				//fmt.Println("before")
				//tx.PrintTx()
				if blc.VerifyTransaction(&tx, _txs) {

					txs = append(txs, &tx)
				}
			}

			if len(txs) == 1 {

				fmt.Println("All transactions invalid!\n")

			}

			fmt.Println("All transactions verified succ!\n")


			// 建立新区块
			var block *Block
			// 取出上一个区块
			err = blc.DB.View(func(tx *bolt.Tx) error {

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

			//构造新区块
			block = NewBlock(txs, block.Height+1, block.Hash)

			fmt.Println("New block is mined!")

			// 添加到数据库
			err = blc.DB.Update(func(tx *bolt.Tx) error {

				b := tx.Bucket([]byte(blockTableName))
				if b != nil {

					b.Put(block.Hash, block.Serialize())
					b.Put([]byte(newestBlockKey), block.Hash)
					blc.Tip = block.Hash

				}
				return nil
			})
			if err != nil {

				log.Panic(err)
			}

			utxoSet := UTXOSet{blc}
			utxoSet.Update()

			// 去除内存池中打包到区块的交易
			for _, tx := range txs {

				fmt.Println("delete...")
				txHash := hex.EncodeToString(tx.TxHAsh)
				delete(memTxPool, txHash)
			}

			// 发送区块给其他节点
			sendBlock(knowedNodes[0], block.Serialize())
			//for _, node := range knownNodes {
			//	if node != nodeAddress {
			//		sendInv(node, "block", [][]byte{newBlock.Hash})
			//	}
			//}

			if len(memTxPool) > 0 {

				goto MineTransactions
			}
		}
	}
}
```

好累啊，终于将一次网络同步需要通讯的消息类型写完了。是不是觉得好复杂，其实不然，一会结合实际🌰看过程就好理解多了。

## Server服务器端

由于我们是在本地模拟网络环境，所以采用不同的端口号来模拟节点IP地址。eg：localhost:8000代表一个节点，eg：localhost:8001代表一个不同的节点。

写一个启动Server服务的方法：

```

func StartServer(nodeID string, minerAdd string) {

	// 当前节点IP地址
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	// 挖矿节点设置
	if len(minerAdd) > 0 {

		miningAddress = minerAdd
	}

	// 启动网络监听服务
	ln, err := net.Listen(PROTOCOL, nodeAddress)
	if err != nil {

		log.Panic(err)
	}
	defer ln.Close()

	blc := GetBlockchain(nodeID)
	//fmt.Println("startserver\n")
	//blc.Printchain()

	// 第一个终端：端口为3000,启动的就是主节点
	// 第二个终端：端口为3001，钱包节点
	// 第三个终端：端口号为3002，矿工节点
	if nodeAddress != knowedNodes[0] {

		// 该节点不是主节点，钱包节点向主节点请求数据
		sendVersion(knowedNodes[0], blc)
	}

	for {

		// 接收客户端发来的数据
		connc, err := ln.Accept()
		if err != nil {

			log.Panic(err)
		}

		// 不同的命令采取不同的处理方式
		go handleConnection(connc, blc)
	}
}
```

针对不同的命令要采取不同的处理方式(上面已经讲了具体命令对应的实现)，所以需要实现一个命令解析器：

```
// 客户端命令处理器
func handleConnection(conn net.Conn, blc *Blockchain) {

	//fmt.Println("handleConnection:\n")
	//blc.Printchain()

	// 读取客户端发送过来的所有的数据
	request, err := ioutil.ReadAll(conn)
	if err != nil {

		log.Panic(err)
	}

	fmt.Printf("Receive a Message:%s\n", request[:COMMANDLENGTH])

	command := bytesToCommand(request[:COMMANDLENGTH])

	switch command {

	case COMMAND_VERSION:
		handleVersion(request, blc)

	case COMMAND_ADDR:
		handleAddr(request, blc)

	case COMMAND_BLOCK:
		handleBlock(request, blc)

	case COMMAND_GETBLOCKS:
		handleGetblocks(request, blc)

	case COMMAND_GETDATA:
		handleGetData(request, blc)

	case COMMAND_INV:
		handleInv(request, blc)

	case COMMAND_TX:
		handleTx(request, blc)

	default:
		fmt.Println("Unknown command!")
	}
	defer conn.Close()
}
```

Server需要的一些全局变量：

```
//localhost:3000 主节点的地址
var knowedNodes = []string{"localhost:8000"}
var nodeAddress string //全局变量，节点地址
// 存储拥有最新链的未处理的区块hash值
var unslovedHashes [][]byte
// 交易内存池
var memTxPool = make(map[string]Transaction)
// 矿工地址
var miningAddress string
// 挖矿需要满足的最小交易数
const minMinerTxCount = 1
```

为了能使矿工节点执行挖矿的责任，修改启动服务的CLI代码。当带miner参数且不为空时，该参数为矿工奖励地址。

```
startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
flagMiner := startNodeCmd.String("miner","","定义挖矿奖励的地址......")

```
```
func (cli *CLI) startNode(nodeID string, minerAdd string)  {

	fmt.Printf("start Server:localhost:%s\n", nodeID)
	// 挖矿地址判断
	if len(minerAdd) > 0 {

		if IsValidForAddress([]byte(minerAdd)) {

			fmt.Printf("Miner:%s is ready to mining...\n", minerAdd)
		}else {

			fmt.Println("Server address invalid....\n")
			os.Exit(0)
		}
	}

	// 启动服务器
	StartServer(nodeID, minerAdd)
}
```

除此之外，转账的send命令也需要稍作修改。带有mine参数表示立即挖矿，由交易的第一个转账方地址进行挖矿；如果没有该参数，表示由启动服务的矿工进行挖矿。

```
flagSendBlockMine := sendBlockCmd.Bool("mine",false,"是否在当前节点中立即验证....")
```
```
//转账
func (cli *CLI) send(from []string, to []string, amount []string, nodeID string, mineNow bool)  {

	blc := GetBlockchain(nodeID)
	defer blc.DB.Close()

	utxoSet := &UTXOSet{blc}

	// 由交易的第一个转账地址进行打包交易并挖矿
	if mineNow {

		blc.MineNewBlock(from, to, amount, nodeID)

		// 转账成功以后，需要更新UTXOSet
		utxoSet.Update()
	}else {
		
		// 把交易发送到矿工节点去进行验证
		fmt.Println("miner deal with the Tx...")

		// 遍历每一笔转账构造交易
		var txs []*Transaction
		for index, address := range from {

			value, _ := strconv.Atoi(amount[index])
			tx := NewTransaction(address, to[index], int64(value), utxoSet, txs, nodeID)
			txs = append(txs, tx)

			// 将交易发送给主节点
			sendTx(knowedNodes[0], tx)
		}
	}
}
```

# 网络同步🌰详解

假设现在的情况是这样的：

- A节点(中心节点)，拥有3个区块的区块链
- B节点(钱包节点)，拥有1个区块的区块链
- C节点(挖矿节点)，拥有1个区块的区块链

很明显，B节点需要向A节点请求2个区块更新到自己的区块链上。那么，实际的代码逻辑是怎样处理的？

### 中心节点与钱包节点的同步逻辑
A和B都是既可以充当服务端，也可以充当客户端。

> 1. A.StartServer 等待接收其他节点发来的消息

> 2. B.StartServer 启动同步服务

> 3. B != 中心节点，向中心节点发请求:B.sendVersion(A, B.blc)

> 4. A.Handle(B.Versin) :A收到B的Version消息
  > 4.1 A.blc.Height > B.blc.Height(3>1)  A.sendVersion(B, A.blc)

> 5. B.Handle(A.Version):B收到A的Version消息
  5.1 B.blc.Height > A.blc.Height(1<3) B向A请求其所有的区块哈希:B.sendGetBlocks(B)

> 6. A.Handle(B.GetBlocks) A将其所有的区块哈希返回给B:A.sendInv(B, "block",blockHashes)

> 7. B.Handle(A.Inv) B收到A的Inv消息
  7.1取第一个哈希，向A发送一个消息请求该哈希对应的区块:B.sendGetData(A, blockHash)
  7.2在收到的blockHashes去掉请求的blockHash后，缓存到一个数组unslovedHashes中

> 8. A.Handle(B.GetData) A收到B的GetData请求，发现是在请求一个区块
  8.1 A取出对应得区块并发送给B:A.sendBlock(B, block)

> 9. B.Handle(A.Block) B收到A的一个Block
  9.1 B判断该Block自己是否拥有，如果没有加入自己的区块链
  9.2 len(unslovedHashes) != 0，如果还有区块未处理，继续发送GetData消息，相当于回7.1:B.sendGetData(A,unslovedHashes[0])
9.3 len(unslovedHashes) == 0,所有A的区块处理完毕，重置UTXO数据库

>10. 大功告成

### 挖矿节点参与的同步逻辑

上面的同步并没有矿工挖矿的工作，那么由矿工节点参与挖矿时的同步逻辑又是怎样的呢？

> 1. A.StartServer 等待接收其他节点发来的消息

> 2. C.StartServer 启动同步服务，并指定自己为挖矿节点，指定挖矿奖励接收地址

> 3. C != 中心节点，向中心节点发请求:C.sendVersion(A, C.blc)

> 4. A.Handle(C.Version),该步骤如果有更新同上面的分析相同

> 5. B.Send(B, C, amount) B给C的地址转账形成一笔交易
    5.1 B.sendTx(A, tx) B节点将该交易tx转发给主节点做处理
    5.2 A.Handle(B.tx) A节点将其信息分装到Inv发送给其他节点:A.SendInv(others, txInv)

> 6. C.Handle(A.txInv),C收到转发的交易将其放到交易缓冲池memTxPool，当memTxPool内Tx达到一定数量就进行打包挖矿产生新区块并发送给其他节点：C.sendBlock(others, blockData)

> 7. A(B).HandleBlock(C. blockData) A和B都会收到C产生的新区块并添加到自己的区块链上

> 8.大功告成

# 几个命令总结

从上面的🌰可以看出，这几个命令总是两两对应的。而且很明显有些命令用于低高度节点请求，有些命令用于最新链节点对请求节点的回复。

### A & B
还是以上面的情形为例，A为主节点(3个区块高度)，B(1个区块)。这里最开始是由B先发起的Version_B消息。

| 命令对 | A(回复) | B(发起) |
| :------:| :------: | :------: |
|CmdPair0 | Version_A | Version_B |
|CmdPair1 | Inv | GetBlocks |
|CmdPair2 |BlockData/TxData | GetData |
|Action| -- |HandleBlock/HandleTx |

我们将表格添加一些流程走向，就能直观地看出整个区块同步的过程中需要的几次通讯。

![区块同步通讯流程图1.png](https://upload-images.jianshu.io/upload_images/830585-094930c28a66452b.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

### A & C

| CMD | A | B |  C|
| :------:| :------: | :------: | :------: |
|CmdPair0 | Version_A || Version_C |
|(同A&B同步流程) | ... | --  |区块同步到最新 |
| sendAction |--|send(B,C,amount)|-- |
| sendTx | -- | TxData | -- |
| sendInv | txInv |-- | Handle(tx) |
| sendBlock |-- |-- | BlockData |
| handleBlock | Handle(BlockData) | Handle(BlockData) | -- |

![区块同步通讯流程图2.png](https://upload-images.jianshu.io/upload_images/830585-a37d2f26535d743a.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

## 节点设置

我们通过设置一个环境变量NODE_ID来区别不同的节点。通过“export NODE_ID=8888”命令来在终端设置节点，通过以下方式在代码CLI.RUN中获取到节点的端口号：

```
//获取节点
	//在命令行可以通过 export NODE_ID=8888 设置节点ID
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {

		fmt.Printf("NODE_ID env var is not set!\n")
		os.Exit(1)
	}
	fmt.Printf("NODE_ID:%s\n", nodeID)
```

有了节点的概念，在这里为了模拟不同节点的区块链，我们需要给相关方法加入节点作为参数。

例如创建区块链(CreateBlockchainWithGensisBlock)方法中加入节点参数来表示该区块链属于哪一个节点，钱包创建(NewWallets)，交易挖矿(MineNewBlock)等。相应地在CLI中的调用也要做相应的修改。这里只以CreateBlockchainWithGensisBlock为例，其他参照源代码修改下。

修改数据库名字宏定义(钱包文件名同理)

```
//相关数据库属性
const dbName = "chaorsBlockchain_"
```
创建区块链

```
//1.创建创世区块
func CreateBlockchainWithGensisBlock(address string, nodeID string) *Blockchain {

	//格式化数据库名字，表示该链属于哪一个节点
	dbName := fmt.Sprintf(dbName, nodeID)
        ……
        .......
}
```

CLI调用

```
//新建区块链
func (cli *CLI)creatBlockchain(address string, nodeID string)  {

	blockchain := CreateBlockchainWithGensisBlock(address, nodeID)
	defer blockchain.DB.Close()
}
```

# 撸起袖子就是干

主节点：8000
钱包节点：8001
矿工节点：8002

打开终端1

```
// 1.设置节点端口为8000
export NODE_ID=8000

// 2.编译项目
go build main.go

// 3.创建钱包
./main createWallet

// 4.创建区块链
./main createBlockchain -address

// 5.备份创世区块链(因为后面要改变这个区块链)
cp chaorsBlockchain_8000.db chaorsBlockchain_genesis.db
```

打开终端2 

```
// 6.设置节点端口为8001
export NODE_ID=8001

// 7.创建两个钱包地址
./main creatWallet
./main creatWallet
```

切换到终端1

```
// 8. 进行两次转账 额度分别为22，11
./main send -from...  -mine

// 9.启动同步服务
./main startnode
```

切换到终端2

```
// 10.启动同步服务
./main startnode

// 11.查询余额
./main getBalance -address
```

![Node8000_0.png](https://upload-images.jianshu.io/upload_images/830585-995dd1a4a0786d75.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

![Node8001_0.png](https://upload-images.jianshu.io/upload_images/830585-3e109680e6280ff8.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

切换到8000(终端1)，8002(终端3)

```
// 12. 启动8001，8002节点网络服务。并将8002节点设置为矿工节点
//8000
./main startnode
//8002
./main startnode -miner
```

切换到8001(终端2)

```
// 13.从8001的钱包给8002转账11
send -from ... 

// 14.启动节点同步服务
./main startnode 
```

![Node8000_1.png](https://upload-images.jianshu.io/upload_images/830585-6c525b9f62924d79.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

![Node8001_1.png](https://upload-images.jianshu.io/upload_images/830585-08df9360e5d36bb3.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

![Node8002.png](https://upload-images.jianshu.io/upload_images/830585-92be5350297f227c.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

太坎坷了，今天终于把网络同步的笔记写完了。从写代码到Debug，再到笔记成稿，说多了都是泪啊。

这次代码修改了之前交易签名和验签时候的代码，具体就不多说了，详见源码吧。














