package main

import (

	"chaors.com/LearnGo/publicChaorsChain/part5-cli-Prototype/BLC"
	//"fmt"
	//"os"
	//"flag"
	//"log"
)

func main() {


	//fmt.Printf("%v\n", BLC.IsDBExists("chaorsBlockchain.db"))
	blockchain := BLC.CreateBlockchainWithGensisBlock()
	//fmt.Printf("%v\n", BLC.IsDBExists("chaorsBlockchain.db"))

	//blockchain.Printchain()
	cli := BLC.CLI{blockchain}
	cli.Run()


	//defer blockchain.DB.Close()


	/**
	addBlockCmd := flag.NewFlagSet("addBlock", flag.ExitOnError)
	printchainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	flagAddBlockData := addBlockCmd.String("data", "chaors", "交易数据")
	switch os.Args[1] {
	case "addBlock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {

		if *flagAddBlockData == "" {

			printUsage()
			os.Exit(1)
		}

		fmt.Println(*flagAddBlockData)
	}
	if printchainCmd.Parsed() {

		fmt.Println("输出所有区块信息")
	}
	*/


	//args := os.Args
	//fmt.Printf("%v\n", args)

	//定义一个字符串变量，并制定默认值以及使用方式
	//flagStr := flag.String("printchain", "", "输出所有区块信息")
	//
	////定义一个int型字符 默认值为6
	//flagInt := flag.Int("number", 6, "输出一个数")
	//
	////定义一个bool型变量 默认值为false
	//flagBool := flag.Bool("open", false, "判断真假")
	//
	////flag解析
	//flag.Parse()
	//
	//fmt.Printf("%s\n", *flagStr)
	//fmt.Printf("%d\n", *flagInt)
	//fmt.Printf("%v\n", *flagBool)


	//
	////添加一个新区快
	//blockchain.AddBlockToBlockchain("first Block")
	//blockchain.AddBlockToBlockchain("second Block")
	//blockchain.AddBlockToBlockchain("third Block")
	//
	//blockchain.Printchain()

}
