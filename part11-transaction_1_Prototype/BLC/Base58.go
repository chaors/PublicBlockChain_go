package BLC

import (
	"math/big"
	"bytes"
)

/**
Base58是用于比特币中使用的一种独特的编码方式，主要用于产生比特币的钱包地址。

相比的Base64，Base58不使用数字 “0”，字母大写 “O”，字母大写 “I”，和字母小写 “L”，以及 “+” 和 “/” 符号。

设计Base58主要的目的是：

避免混淆。在某些字体下，数字0和字母大写O，以及字母大写我和字母小写升会非常相似。
不使用 “+” 和 “/” 的原因是非字母或数字的字符串作为帐号较难被接受。
没有标点符号，通常不会被从中间分行。
大部分的软件支持双击选择整个字符串。
但是这个base58的计算量比BASE64的计算量多了很多。因为58不是2的整数倍，需要不断用除法去计算。
而且长度也比的base64稍微多了一点。
*/

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

