package BLC

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"bytes"
)

//用于生成地址的版本
const AddVersion  = byte(0x00)
//用于生成地址的校验和位数
const AddressChecksumLen = 4


type Wallet struct {
	//私钥
	PrivateKey ecdsa.PrivateKey
	//公钥
	PublicKey []byte
}

//1.创建钱包
func NewWallet () *Wallet  {

	privateKey, publicKey := newKeyPair()

	return &Wallet{privateKey, publicKey}
}

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

//2.获取钱包地址 根据公钥生成地址
func (wallet *Wallet) GetAddress() []byte {

	//1.使用RIPEMD160(SHA256(PubKey)) 哈希算法，取公钥并对其哈希两次
	ripemd160Hash := Ripemd160Hash(wallet.PublicKey)
	//2.拼接版本
	version_ripemd160Hash := append([]byte{AddVersion}, ripemd160Hash...)
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
func  IsValidForAddress(address []byte) bool {

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