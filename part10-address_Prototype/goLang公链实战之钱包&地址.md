# goLang公链实战之钱包&地址

之前我们的项目中转账什么的都是使用的字符串做用户名，但是在比特币种并没有用户账户的概念。所有的交易都是基于地址进行转账的，所谓的地址本质是一个公钥，地址只是把公钥用人们可读的方式表现出来。

第一个比特币地址：[1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa](https://blockchain.info/address/1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa)


![第一个比特币地址.png](https://upload-images.jianshu.io/upload_images/830585-142b9089ef747193.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

地址是基于一系列加密算法完成的，今天将循序渐进地学习这些算法和相关基础方法的实现。

### Base58加密

Base58是用于比特币中使用的一种独特的编码方式，主要用于产生比特币的钱包地址。相比的Base64，Base58不使用数字 “0”，字母大写 “O”，字母大写 “I”，和字母小写 “L”，以及 “+” 和 “/” 符号。

设计Base58主要的目的是：避免混淆。在某些字体下，数字0和字母大写O，以及字母大写我和字母小写升会非常相似。
不使用 “+” 和 “/” 的原因是非字母或数字的字符串作为帐号较难被接受。

但是这个base58的计算量比BASE64的计算量多了很多。因为58不是2的整数倍，需要不断用除法去计算。而且长度也比的base64稍微多了一点。

```
//base64:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/
//base58:去掉0(零)，O(大写的 o)，I(大写的i)，l(小写的 L)，+，/

//base58编码集
var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// 字节数组转 Base58,加密
func Base58Encode(input []byte) []byte {

	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {

		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	ReverseBytes(result)
	for b := range input {

		if b == 0x00 {

			result = append([]byte{b58Alphabet[0]}, result...)
		} else {

			break
		}
	}

	return result
}

// Base58转字节数组，解密
func Base58Decode(input []byte) []byte {

	result := big.NewInt(0)
	zeroBytes := 0

	for b := range input {

		if b == 0x00 {

			zeroBytes++
		}
	}

	payload := input[zeroBytes:]
	for _, b := range payload {

		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()
	//decoded...表示将decoded所有字节追加
	decoded = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), decoded...)

	return decoded
}
```
# 钱包

### PublicKey-PrivateKey

就像我们生活中我们的纸币往往会放在自己的钱包里，比特币也是如此，每一个钱包有一个地址作为唯一标示。而这个地址是由公钥经过几次哈希算法再经过Base58编码转化而成的。

公钥是由私钥产生的，公私钥总是成对出现的。公钥不是敏感信息，可以告诉其他人。从本质上讲，我们安装一个比特币钱包应用，其实是从比特币客户端生成一个公私钥对，私钥代表你对该钱包的控制权，公钥代表该钱包的地址。

### 钱包结构
```
//1.创建钱包
func NewWallet () *Wallet  {

	privateKey, publicKey := newKeyPair()

	//fmt.Println(privateKey, "\n\n", publicKey)

	return &Wallet{privateKey, publicKey}
}
```
通过私钥创建公钥
```

//通过私钥创建公钥
func newKeyPair() (ecdsa.PrivateKey, []byte) {

	//1.椭圆曲线算法生成私钥
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {

		log.Panic(err)
	}

	//2.通过私钥生成公钥
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)


	return *privateKey, publicKey
}
```

### Address生成

从公钥得到一个地址需要五步走：

##### 1.公钥讲过两次哈希(SHA256+RIPEMD160)得到一个字节数组PubKeyHash
##### 2.PubKeyHash+交易版本Version拼接成一个新的字节数组Version_PubKeyHash
##### 3.对Version_PubKeyHash进行两次哈希(SHA256)并按照一定规则生成校验和CheckSum
##### 4.Version_PubKeyHash+CheckSum拼接成Version_PubKeyHash_CheckSum字节数组
##### 5.对Version_PubKeyHash_CheckSum进行Base58编码即可得到地址Address

通过上面五个步骤，我们知道Address讲过Base58解码后由三部分组成：
交易版本 | 公钥哈希 | 校验和 
- | :-: | -
[Version](https://bitcoin.org/en/developer-reference#raw-transaction-format)  | PubKeyHash | Checksum
0x00  | 62E907B15CBF27D5425399EBF6F0FB50EBB88F18 | C29B7D93

```
//用于生成地址的版本
const Version  = byte(0x00)
//用于生成地址的校验和位数
const AddressChecksumLen = 4

//2.获取钱包地址 根据公钥生成地址
func (wallet *Wallet) GetAddress() []byte {

	//1.使用RIPEMD160(SHA256(PubKey)) 哈希算法，取公钥并对其哈希两次
	ripemd160Hash := Ripemd160Hash(wallet.PublicKey)
	//2.拼接版本
	version_ripemd160Hash := append([]byte{Version}, ripemd160Hash...)
	//3.两次sha256生成校验和
	checkSumBytes := CheckSum(version_ripemd160Hash)
	//4.拼接校验和
	bytes := append(version_ripemd160Hash, checkSumBytes...)

	//5.base58编码
	return Base58Encode(bytes)
}

//将公钥进行两次哈希
func Ripemd160Hash(publicKey []byte) []byte  {

	//1.hash256
	hash256 := sha256.New()
	hash256.Write(publicKey)
	hash := hash256.Sum(nil)

	//2.ripemd160
	ripemd160 := ripemd160.New()
	ripemd160.Write(hash)

	return ripemd160.Sum(nil)
}

//两次sha256哈希生成校验和
func CheckSum(bytes []byte) []byte {

	//hasher := sha256.New()
	//hasher.Write(bytes)
	//hash := hasher.Sum(nil)
	//与下面一句等同
	//hash := sha256.Sum256(bytes)

	hash1 := sha256.Sum256(bytes)
	hash2 := sha256.Sum256(hash1[:])

	return hash2[:AddressChecksumLen]
}

//3.判断地址是否有效
func (wallet *Wallet) IsValidForAddress(address []byte) bool {

	//1.base58解码地址得到版本，公钥哈希和校验位拼接的字节数组
	version_publicKey_checksumBytes := Base58Decode(address)
	//2.获取校验位和version_publicKeHash
	checkSumBytes := version_publicKey_checksumBytes[len(version_publicKey_checksumBytes)-AddressChecksumLen:]
	version_ripemd160 := version_publicKey_checksumBytes[:len(version_publicKey_checksumBytes)-AddressChecksumLen]

	//3.重新用解码后的version_ripemd160获得校验和
	checkSumBytesNew := CheckSum(version_ripemd160)

	//4.比较解码生成的校验和CheckSum重新计算的校验和
	if bytes.Compare(checkSumBytes, checkSumBytesNew) == 0 {

		return true
	}

	return false
}
```

至此，就可以得到一个**真实的比特币地址**，你甚至可以在 [blockchain.info](https://blockchain.info/) 查看它的余额。不过我可以负责任地说，无论生成一个新的地址多少次，检查它的余额都是 0。

## 钱包集Wallets

有了钱包后，链上所有钱包需要一个统一的管理，这就有赖于Wallets类。当我们在比特币上创建一个钱包的地址后，那么这个钱包的地址将一直存在。因此，Wallets钱包集也需要实现数据持久化。这里我们采用文件去存储钱包集。

#### 创建Wallets

```
//存储钱包集的文件名
const WalletFile = "Wallets.dat"

type Wallets struct {
	Wallets map[string] *Wallet
}

//1.创建钱包集合
func NewWallets() (*Wallets, error) {

	//判断文件是否存在
	if _, err := os.Stat(WalletFile); os.IsNotExist(err) {

		wallets := &Wallets{}
		wallets.Wallets = make(map[string] *Wallet)

		return wallets, err
	}


	var wallets Wallets
	//读取文件
	fileContent, err := ioutil.ReadFile(WalletFile)
	if err != nil {

		log.Panic(err)
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {

		log.Panic(err)
	}

	return &wallets, err
}
```

#### 创建Wallet

```
//2.创建新钱包
func (wallets *Wallets) CreateWallet()  {

	wallet := NewWallet()
	fmt.Printf("Your new addres：%s\n",wallet.GetAddress())
	wallets.Wallets[string(wallet.GetAddress())] = wallet

	//保存到本地
	wallets.SaveWallets()
}

//3.保存钱包集信息到文件
func (wallets *Wallets) SaveWallets()  {

	var context bytes.Buffer

	//注册是为了可以序列化任何类型
	gob.Register(elliptic.P256())
	encoder :=gob.NewEncoder(&context)
	err := encoder.Encode(&wallets)
	if err != nil {

		log.Panic(err)
	}

	// 将序列化以后的数覆盖写入到文件
	err = ioutil.WriteFile(WalletFile, context.Bytes(), 0664)
	if err != nil {

		log.Panic(err)
	}
}
```

## Wallet创建集成到CLI工具

#### createWallet
```
//打印目前左右命令使用方法
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tcreateBlockchain -address --创世区块地址 ")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT --交易明细")
	fmt.Println("\tprintchain --打印所有区块信息")
	fmt.Println("\tgetbalance -address -- 输出区块信息.")
	fmt.Println("\tcreateWallet -- 创建钱包.")
	fmt.Println("\tgetAddressList -- 输出所有钱包地址.")
}
```

新增createWallet命令和相关解析：

```
//命令创建
createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)

//输入命令判断
switch os.Args[1] {
	...
        case "createWallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
}

//命令解析
//创建钱包
if createWalletCmd.Parsed() {

        cli.createWallet()
}
```

新增CLI_createWallet.go文件，真正地实现钱包创建功能：
```
func (cli *CLI)createWallet()  {

	wallets, _ := NewWallets()
	wallets.CreateWallet()

	fmt.Println(len(wallets.Wallets))
}
```

#### getAddressList

该命令用于获取当前区块链上已创建的所有钱包地址。

```
getAddressListCmd := flag.NewFlagSet("getAddressList", flag.ExitOnError)

switch os.Args[1] {
	...
        case "getAddressList":
		err := getAddressListCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
}

//获取所有钱包地址
	if getAddressListCmd.Parsed() {

		cli.getAddressList()
	}
```

CLI_getAddressList.go
```
func (cli *CLI) getAddressList()  {

	fmt.Println("All addresses:")

	wallets, _ := NewWallets()
	for address, _ := range wallets.Wallets {

		fmt.Println(address)
	}
}
```

# Transaction集成Address

我们现在已经有了地址的概念，所以之前使用用户名做转账依据的地方都可以集成Address。

#### CLI_send命令

当我们进行一笔转账时，需要先对钱包地址进行判断是否有效。在地址有效的前提下，才能进行后续的转账操作。

```
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
		from := Json2Array(*flagSendBlockFrom)
		to := Json2Array(*flagSendBlockTo)

		//输入地址有效性判断
		for index, fromAddress := range from {

			if BLC.IsValidForAdress([]byte(fromAddress)) == false || BLC.IsValidForAdress([]byte(to[index])) == false {

				fmt.Printf("地址%s无效", fromAddress)
				os.Exit(1)
			}
		}

		amount := Json2Array(*flagSendBlockAmount)

		cli.send(from, to, amount)
	}
```

#### TXInput

之前交易输入的结构是这样的：
```
type TXInput struct {
	//交易ID
	TxHash []byte
	//存储TXOutput在Vouts里的索引
	Vout int
	//数字签名
	ScriptSig string
}
```

由于引入了地址和公私钥的概念，这里就可以给交易输入引入签名和公钥属性。这里且不论什么是签名，公钥代表这笔输入属于哪一个钱包。
```

type TXInput struct {
	//交易ID
	TxHash []byte
	//存储TXOutput在Vouts里的索引
	Vout int
	//数字签名
	Signature []byte
	//公钥
	PublicKey []byte
}
```

之前我们验证交易输入是否属于一个账户时，由于我们设定的账户值和公钥直接是一个字符串形式的用户名，直接比较即可。现在输入的是一个地址，又该怎么办呢？

我们知道地址是由公钥进行多次哈希和按一定规则运算得出的，由于哈希是不可逆的，我们不可能根据地址反推出公钥然后和交易输入的公钥属性去作比较。这个时候，就需要拿地址Base58解码后得到的公钥哈希去和拿按交易输入的公钥进行两次哈希得到的值进行比较即可。

```
//验证当前输入是否是当前地址的
func (txInput *TXInput) UnlockWithAddress(address string) bool  {

	//base58解码
	version_pubKeyHash_checkSumBytes := Base58Decode([]byte(address))
	//去除版本得到地反编码的公钥两次哈希后的值
	ripemd160Hash := version_pubKeyHash_checkSumBytes[1:len(version_pubKeyHash_checkSumBytes)-4]

	//Ripemd160Hash算法得到公钥两次哈希后的值
	ripemd160HashNew := Ripemd160Hash(txInput.PublicKey)

	return bytes.Compare(ripemd160HashNew, ripemd160Hash) == 0
}
```

#### TXOutput

我们知道比特币种对每一笔交易中的输出都会做一个锁定，将其锁定为某一个钱包所拥有；当这笔交易输出用于下一笔交易作为交易输入时需要进行解锁操作以保障花费的这笔TXOutput是属于当前转账方钱包的。

引入钱包的概念后，我们就可以实现TXOutput的锁定和解锁。

```
type TXOutput struct {
	//面值
	Value int64
	//用户名
	Ripemd160Hash []byte  //用户名  公钥两次哈希后的值
}

func NewTXOutput(value int64,address string) *TXOutput {

	txOutput := &TXOutput{value,nil}

	// 设置Ripemd160Hash
	txOutput.Lock(address)

	return txOutput
}

//锁定
func (txOutput *TXOutput) Lock(address string) {

	version_pubKeyHash_checkSumBytes := Base58Decode([]byte(address))
	txOutput.Ripemd160Hash = version_pubKeyHash_checkSumBytes[1:len(version_pubKeyHash_checkSumBytes)-4]
}

//解锁
func (txOutput *TXOutput) UnLockScriptPubKeyWithAddress(address string) bool {

	version_pubKeyHash_checkSumBytes := Base58Decode([]byte(address))
	ripemd160Hash := version_pubKeyHash_checkSumBytes[1:len(version_pubKeyHash_checkSumBytes) - 4]

	//fmt.Println(txOutput.Ripemd160Hash, ripemd160Hash)
	return bytes.Compare(txOutput.Ripemd160Hash, ripemd160Hash) == 0
}
```

# 数字签名

通俗地讲，数字签名就是每一笔交易的证明。如果一个交易的数字签名是无效的，那么这笔交易就会被认为是无效的，因此，这笔交易也就无法被加到区块链中。

数字签名的主要作用有两点：其一，证明该交易是转账方发起的。其二，证明交易信息没有被更改。

当某个地址发起一笔转账时，需要首先打包成交易并对该交易进行数字摘要组成一段字符串，然后再用自己的私钥对摘要字符串进行加密形成数字签名。发起转账的用户会把交易信息和数字签名一起发送给矿工，矿工会用转账用户的公钥对数字签名进行验签，如果验证成功说明交易确实是该转账方发起的且交易信息未被篡改，交易有效可以打包到区块内。

从前面的学习中，我们知道一笔交易包含交易哈希可以证明整个交易信息是否被篡改；交易输入引用的之前交易产生的UTXO可以证明交易发起方是谁；因此，我们签名的内容包括交易的交易哈希和交易输入里引用的TXOutput的公钥哈希。

签名的产生依赖于EDSA椭圆曲线加密算法，通过私钥PrivateKey经过一定运算得到签名。

由于在签名过程中，我们的目的是得到交易的签名。因此为了保证交易其他信息不被改变，我们需要在计算过程中对交易进行一个拷贝。由于计算前签名为空，签名的过程并不需要交易输入的公钥值。因此，拷贝的交易的签名和公钥置为nil。

由于创币交易的特殊性(其没有交易输入)，所以创币交易不需要进行签名。

#### 废话少说上代码

###### 交易签名

```
//数字签名
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTxs map[string]Transaction) {

	//判断当前交易是否为创币交易，coinbase交易因为没有实际输入，所以没有被签名
	if tx.IsCoinbaseTransaction() {

		return
	}

	for _, vin := range tx.Vins {

		if prevTxs[hex.EncodeToString(vin.TxHash)].TxHAsh == nil {

			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	//将会被签署的是修剪后的交易副本
	txCopy := tx.TrimmedCopy()

	//遍历交易的每一个输入
	for inID, vin := range txCopy.Vins {

		//交易输入引用的上一笔交易
		prevTx := prevTxs[hex.EncodeToString(vin.TxHash)]
		//Signature 被设置为 nil
		txCopy.Vins[inID].Signature = nil
		//PubKey 被设置为所引用输出的PubKeyHash
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
		//设置交易哈希
		txCopy.TxHAsh = txCopy.Hash()
		//设置完哈希后要重置PublicKey
		txCopy.Vins[inID].PublicKey = nil

		// 签名代码
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.TxHAsh)
		if err != nil {

			log.Panic(err)
		}
		//一个ECDSA签名就是一对数字，我们对这对数字连接起来就是signature
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vins[inID].Signature = signature
	}
}
```
###### 交易拷贝
```
// 拷贝一份新的Transaction用于签名,包含所有的输入输出，但TXInput.Signature 和 TXIput.PubKey 被设置为 nil                                 T
func (tx *Transaction) TrimmedCopy() Transaction {

	var inputs []*TXInput
	var outputs []*TXOutput

	for _, vin := range tx.Vins {

		inputs = append(inputs, &TXInput{vin.TxHash, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vouts {

		outputs = append(outputs, &TXOutput{vout.Value, vout.Ripemd160Hash})
	}

	txCopy := Transaction{tx.TxHAsh, inputs, outputs}

	return txCopy
}

//对交易信息进行哈希
func (tx *Transaction) Hash() []byte {

	txCopy := tx

	txCopy.TxHAsh = []byte{}

	hash := sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

//交易序列化
func (tx *Transaction) Serialize() []byte {

	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {

		log.Panic(err)
	}

	return encoded.Bytes()
}
```
###### 交易验签

既然有交易签名，那么在打包交易到区块时就需要对交易进行签名验证。交易验签的前半部分和交易签名，不同的是后部分验签通过对签名计算出一对数字，然后利用公钥通过EDSA构造出一个ecdsa.PublicKey，最后借助EDSA加密算法的验证功能去验证构造的公钥，签名对应的一堆数字以及交易哈希。验证通过，说明交易验签成功。

```
// 验签
func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {

	if tx.IsCoinbaseTransaction() {

		return true
	}

	for _, vin := range tx.Vins {

		if prevTxs[hex.EncodeToString(vin.TxHash)].TxHAsh == nil {

			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	//用于椭圆曲线算法生成秘钥对
	curve := elliptic.P256()

	// 遍历输入，验证签名
	for inID, vin := range tx.Vins {

		// 这个部分跟Sign方法一样,因为在验证阶段,我们需要的是与签名相同的数据。
		prevTx := prevTxs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHAsh = txCopy.Hash()
		txCopy.Vins[inID].PublicKey = nil

		// 私钥
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		// 公钥
		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.PublicKey[(keyLen / 2):])

		// 使用从输入提取的公钥创建ecdsa.PublicKey
		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.TxHAsh, &r, &s) == false {

			return false
		}
	}

	return true
}
```
## 签名的集成

对交易签名首先需要某个方法能获取到某个交易。
```
//获取某个交易
func (blc *Blockchain) FindTransaction(txHash []byte) (Transaction, error) {

	blcIterator := blc.Iterator()

	for {

		block := blcIterator.Next()

		for _, tx := range block.Txs {

			if bytes.Compare(tx.TxHAsh, txHash) == 0 {

				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {

			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}
```
区块链层级的交易签名和验签函数

```
//交易签名
func (blc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {

	if tx.IsCoinbaseTransaction() {

		return
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vins {

		//找到当前交易输入引用的所有交易
		prevTX, err := blc.FindTransaction(vin.TxHash)
		if err != nil {

			log.Panic(err)
		}

		prevTXs[hex.EncodeToString(prevTX.TxHAsh)] = prevTX
	}

	tx.Sign(privKey, prevTXs)

}

// 交易验签
func (blc *Blockchain) VerifyTransaction(tx *Transaction) bool {

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vins {

		prevTX, err := blc.FindTransaction(vin.TxHash)
		if err != nil {

			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.TxHAsh)] = prevTX
	}

	return tx.Verify(prevTXs)
}
```

签名发生在构造交易时，验签发生在交易打包到区块之前。所以需要在Transaction类里构造交易时集成签名，在Blockchain类里的挖矿函数集成验证签名。

Transaction.NewTransaction
```
//2.普通交易
func NewTransaction(from string, to string, amount int, blc *Blockchain, txs []*Transaction) *Transaction {
      ...
      ...

      tx.HashTransactions()

	  //进行签名
	  blc.SignTransaction(tx, wallet.PrivateKey)

	  return tx
}
```

Blockchain.Mine
```
//2.新增一个区块到区块链 --> 包含交易的挖矿
func (blc *Blockchain) MineNewBlock(from []string, to []string, amount []string) {

	var txs []*Transaction
        ...
        ...
    //建立新区快前需要对交易进行验签，
	for _, tx := range txs {

		if blc.VerifyTransaction(tx) == false {

			log.Printf("The Tx:s% verify failed.", tx.TxHAsh)
		}
	}

	//3.建立新区块
	block = NewBlock(txs, block.Height+1, block.Hash)
        ...
        ...
}
```

至此，已经实现了地址的定义并集成到项目中。项目的功能虽然保持没变，但是集成了钱包地址后会更向一个区块链公链。

## Main_Tset
```
package main

import (

	"chaors.com/LearnGo/publicChaorsChain/part10-address_Prototype/BLC"
)

func main() {

	cli := BLC.CLI{}
	cli.Run()

	/**
	//Wallet
	ripemd160.New()
	wallet := BLC.NewWallet()
	address := wallet.GetAddress()

	fmt.Printf("%s\n", address)
	//1CiS8axkfLGQUYaeZsuS2Fpv4nVcd6HQqk 和比特币地址相同，可以blockchaininfo查询余额
	//当然一定为0

	//判断地址有效性
	fmt.Println(wallet.IsValidForAddress(address))
	//修改address
	fmt.Println(wallet.IsValidForAddress([]byte("1CiS8axkfLGQUYaeZsuS2Fpv4nVcd6HQqkk")))

	//Wallets
	wallets := BLC.NewWallets()
	wallets.CreateWallet()
	wallets.CreateWallet()
	wallets.CreateWallet()
	fmt.Println()
	fmt.Println(wallets)

	//blc := BLC.CreateBlockchainWithGensisBlock("chaors")
	//utxos := blc.UnUTXOs("chaors")
	//fmt.Println(utxos)
	*/
}
```
RUN

![main_test1](https://upload-images.jianshu.io/upload_images/830585-e64718a56440d8f4.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

以1NUAQrgwDVo72oZjsMkqu8dLd35A8yWFvE地址为例在[blockchain.info](https://blockchain.info/)

![main_test1_1.png](https://upload-images.jianshu.io/upload_images/830585-4baad3889ba1d831.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

地址合法为比特币地址，说明我们的地址生成方法是有效的。我们将该地址稍作修改最后一位改为S:1NUAQrgwDVo72oZjsMkqu8dLd35A8yWFvS，再去查询：

![main_test1_1.png](https://upload-images.jianshu.io/upload_images/830585-90fcf38305d84711.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

多笔转账亦然，只是将之前的用户名改为算法生成的钱包地址，在此不在赘述。
























