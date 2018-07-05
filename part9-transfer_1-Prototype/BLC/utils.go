/**
@author: chaors

@file:   utils.go

@time:   2018/06/21 22:06

@desc:   一些常用的辅助方法
*/


package BLC

import (
	"bytes"
	"encoding/binary"
	"log"
	"encoding/json"
)

//将int64转换为bytes
func IntToHex(num int64) []byte  {

	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {

		log.Panic(err)
	}

	return buff.Bytes()
}

// 标准的JSON字符串转数组
func Json2Array(jsonString string) []string {

	//json 到 []string
	var sArr []string
	if err := json.Unmarshal([]byte(jsonString), &sArr); err != nil {

		log.Panic(err)
	}
	return sArr
}