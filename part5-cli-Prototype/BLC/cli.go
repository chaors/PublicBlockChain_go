package BLC

import (
	"fmt"
	"flag"
	"os"
	"log"
)

type CLI struct {
	//添加cli命令行工具的类
	Blockchain *Blockchain
}

//打印目前左右命令使用方法
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tcreateBlockchainWithGensisBlock -data Data 创世信息  创建创世区块")
	fmt.Println("\taddBlock -data DATA 交易信息   新增区块")
	fmt.Println("\tprintchain 打印所有区块信息")
}

func isValidArgs() {

	//获取当前输入参数个数
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

//新增区块
func (cli *CLI) addBlock(data string) {
	cli.Blockchain.AddBlockToBlockchain(data)

}

//打印区块链
func (cli *CLI) printchain() {
	cli.Blockchain.Printchain()
}

func (cli *CLI) Run() {

	isValidArgs()

	//自定义cli命令
	addBlockCmd := flag.NewFlagSet("addBlock", flag.ExitOnError)
	printchainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	//addBlockCmd 设置默认参数
	flagAddBlockData := addBlockCmd.String("data", "chaors", "交易数据")

	//解析输入的第二个参数是addBlock还是printchain，第一个参数为./main
	switch os.Args[1] {
	case "addBlock":
		//第二个参数为相应命令，取第三个参数开始作为参数并解析
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1)
	}

	//对addBlockCmd命令的解析
	if addBlockCmd.Parsed() {

		if *flagAddBlockData == "" {

			printUsage()
			os.Exit(1)
		}

		//这里真正地调用新增区块方法
		cli.addBlock(*flagAddBlockData)
	}
	//对printchainCmd命令的解析
	if printchainCmd.Parsed() {

		cli.printchain()
	}
}
