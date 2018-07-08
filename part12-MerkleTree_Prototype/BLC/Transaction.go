package BLC

import (
	"crypto/sha256"
	"bytes"
	"encoding/gob"
	"log"
	"encoding/hex"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
	"math/big"
	"time"
)

type Transaction struct {
	//1.交易哈希值
	TxHAsh []byte
	//2.输入
	Vins []*TXInput
	//3.输出
	Vouts []*TXOutput
}

//1.coinbaseTransaction
//2.转账时产生的Transaction

/**

 举个简单的🌰，我们先把复杂问题简单化，假设每个区块里只有一个交易。
 1.节点chaors挖到创世区块，产生25BTC的创币交易。由于是创世区块，其本身产生之前是没有
 交易的，所以在输入对象TXInput的交易哈希为空，vount所在的下标为-1，数字签名为空或者
 随便填写；输出对象里btc拥有者为chaors，面值为25btc

 创世区块交易结构
 txInput0 = &TXInput{[]byte{},-1,"Gensis Block"}
 txOutput0 = &TXOutput{25, "chaors"}  索引为0

 Transaction{"00000",
			[]*TXInput{txInput0},
			[]*TXOutput{txOutput0}
			}

 2.chaors获得25btc后，他的好友ww知道后向他索要10btc.大方的chaors便把10btc转给ww.此时
 交易的输入为chaors上笔交易获得的btc,TXInput对象的交易ID为奖励chaors的上一个交易ID，vount下标
 为chaors的TXOutput下标，签名此时且认为是来自chaors，填作"chaors"
 此时chaors的25btc面值的TXOutput就被花费不复存在了，那么chaors还应该有15btc的找零哪去了？
 系统会为chaors的找零新生成一个面值15btc的TXOutput。所以，这次有一个输入，两个输出。

 第二个区块交易结构(假设只有一笔交易)
 chaors(25) 给 ww 转 10 -- >>  chaors(15) + ww(10)

 输入
 txInput1 = &TXInput{"00000",0,"chaors"}
 "00000" 相当于来自于哈希为"00000"的交易
 索引为零，相当于上一次的txOutput0为输入

 输出
 txOutput1 = &TXOutput{10, "ww"}		索引为1  chaors转给ww的10btc产生的输出
 txOutput2 = &TXOutput{15, "chaors"}    索引为2  给ww转账产生的找零
 Transaction{"11111"，
			[]*TXInput{txInput1}
			[]*TXOutput{txOutput1, txOutput2}
			}

 3.ww感觉拥有比特币是一件很酷的事情，又来跟chaors要。出于兄弟情谊，chaors又转给ww7btc

 第三个区块交易结构
 输入
 txInput2 = &TXInput{"11111",2,"chaors"}

 输出
 txOutput3 = &TXOutput{7, "ww"}		  索引为3
 txOutput4 = &TXOutput{8, "chaors"}   索引为4
 Transaction{"22222"，
			[]*TXInput{txInput2}
			[]*TXOutput{txOutput3, txOutput4}
			}

 4.消息传到他们共同的朋友xyz那里，xyz觉得btc很好玩向ww索要15btc.ww一向害怕xyx，于是
 尽管不愿意也只能屈服。我们来看看ww此时的全部财产：
    txOutput1 = &TXOutput{10, "ww"}		索引为1   来自交易"11111"
	txOutput3 = &TXOutput{7, "ww"}		索引为3   来自交易"22222"
 想要转账15btc,ww的哪一笔txOutput都不够，这个时候就需要用ww的两个txOutput都作为
 输入：

 	txInput3 = &TXInput{"11111",1,"ww"}
	txInput4 = &TXInput{"22222",3,"ww"}


 输出
 txOutput5 = &TXOutput{15, "xyz"}		索引为5
 txOutput6 = &TXOutput{2, "ww"}        索引为6

 第四个区块交易结构
 Transaction{"33333"，
			[]*TXInput{txInput3, txInput4}
			[]*TXOutput{txOutput5, txOutput6}
			}

 经过以上交易，chaors最后只剩下面值为8的TXOutput4，txOutput0和txOutput2都在给ww
 的转账中花费；ww最后只剩下面值为2的txOutput6,txOutput3和txOutput4在给xyx的转账
 中花费。由此可见，区块链转账中的UTXO，只要发生交易就不复存在，只会形成新的UTXO
 给新的地址；如果有找零，会产生新的UTXO给原有地址。
*/

//1.创币交易
func NewCoinbaseTransaction(address string) *Transaction {

	//输入  由于创世区块其实没有输入，所以交易哈希传空，TXOutput索引传-1，签名随你
	txInput := &TXInput{[]byte{}, -1, []byte{}, []byte{}}
	//输出  产生一笔奖励给挖矿者
	txOutput := NewTXOutput(int64(25), address)
	txCoinbase := &Transaction{
		[]byte{},
		[]*TXInput{txInput},
		[]*TXOutput{txOutput},
	}

	txCoinbase.HashTransactions()

	return txCoinbase
}

//创币交易判断
func (tx *Transaction) IsCoinbaseTransaction() bool {

	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}

//2.普通交易
func NewTransaction(from string, to string, amount int64, utxoSet *UTXOSet, txs []*Transaction) *Transaction {

	//获取钱包集合
	wallets, _ := NewWallets()
	wallet := wallets.Wallets[from]

	money, spendableUTXODic := utxoSet.FindSpendableUTXOs(from, amount, txs)

	//输入输出
	var txInputs []*TXInput
	var txOutputs []*TXOutput

	for txHash, indexArr := range spendableUTXODic {

		//字符串转换为[]byte
		txHashBytes, _ := hex.DecodeString(txHash)
		for _, index := range indexArr {

			//交易输入
			txInput := &TXInput{
				txHashBytes,
				index,
				nil,
				wallet.PublicKey,
			}

			txInputs = append(txInputs, txInput)
		}
	}

	//转账
	txOutput := NewTXOutput(int64(amount), to)
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = NewTXOutput(int64(money)-int64(amount), from)
	txOutputs = append(txOutputs, txOutput)

	//交易构造
	tx := &Transaction{
		[]byte{},
		txInputs,
		txOutputs,
	}

	tx.HashTransactions()

	//进行签名
	utxoSet.Blockchain.SignTransaction(tx, wallet.PrivateKey, txs)

	return tx

	/**
	//单笔交易构造假数据测试交易
	//输入输出
	var txInputs []*TXInput
	var txOutputs []*TXOutput

	//输入
	txHash, _ := hex.DecodeString("9da0c4b3f53d87f86c496643aacac49f3317ba40424b0ce2410fc174055fc591")
	txInput := &TXInput {
		txHash,
		0,
		from[0],
	}

	fmt.Printf("111--%x\n", txInput.TxHash)

	txInputs = append(txInputs, txInput)

	//转账
	txOutput := &TXOutput{
		4,
	to[0],
	}
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = &TXOutput{
		25-4,
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
	*/

	//1. 有一个函数，返回from这个人所有的未花费交易输出所对应的Transaction
	//unSpentTx := UnSpentTransactionsWithAddress("chaors")
	//fmt.Println(unSpentTx)

	//2. 有一个函数，返回消费的钱 和这些钱的来源的交易的哈希 字典形式
	//TXOutput来自交易的哈希:TXOutput索引的数组
	//	{hash1:[0,2],hash2:[1,4]}

	//return nil
}

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

//将交易信息转换为字节数组
func (tx *Transaction) HashTransactions() {

	//交易信息序列化
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
	if err != nil {

		log.Panic(err)
	}

	//是创币交易的哈希不同
	timeSpBytes := IntToHex(time.Now().Unix())
	//设置hash
	txHash := sha256.Sum256(bytes.Join([][]byte{timeSpBytes, result.Bytes()}, []byte{}))
	tx.TxHAsh = txHash[:]
}
