package BLC

import (
	"fmt"
	"flag"
	"os"
	"log"
)

type CLI struct {

}

//打印目前左右命令使用方法
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tcreateBlockchain -address --创世区块地址 ")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT --交易明细")
	fmt.Println("\tprintchain --打印所有区块信息")
	fmt.Println("\tgetbalance -address -- 输出区块信息.")
	fmt.Println("\tcreateWallet -- 创建钱包.")
	fmt.Println("\tgetAddressList -- 输出所有钱包地址.")
	fmt.Println("\tresetUTXOset -- 测试UTXOSet.")
	fmt.Println("\tstartnode -miner ADDRESS -- 启动节点服务器，并且指定挖矿奖励的地址.")
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

	//获取节点
	//在命令行可以通过 export NODE_ID=8888 设置节点ID
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {

		fmt.Printf("NODE_ID env var is not set!\n")
		os.Exit(1)
	}
	fmt.Printf("NODE_ID:%s\n", nodeID)

	//自定义cli命令
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printchainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createBlockchain", flag.ExitOnError)
	blanceBlockCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)
	getAddressListCmd := flag.NewFlagSet("getAddressList", flag.ExitOnError)
	resetUTXOsetCmd := flag.NewFlagSet("resetUTXOset", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	//addBlockCmd 设置默认参数
	flagSendBlockMine := sendBlockCmd.Bool("mine",false,"是否在当前节点中立即验证....")
	flagSendBlockFrom := sendBlockCmd.String("from", "", "源地址")
	flagSendBlockTo := sendBlockCmd.String("to", "", "目标地址")
	flagSendBlockAmount := sendBlockCmd.String("amount", "", "转账金额")
	flagCreateBlockchainAddress := createBlockchainCmd.String("address", "", "创世区块地址")
	flagBlanceBlockAddress := blanceBlockCmd.String("address", "", "输出区块信息")
	flagMiner := startNodeCmd.String("miner","","定义挖矿奖励的地址......")

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
	case "createWallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getAddressList":
		err := getAddressListCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "resetUTXOset":
		err := resetUTXOsetCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
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
		from := Json2Array(*flagSendBlockFrom)
		to := Json2Array(*flagSendBlockTo)

		//输入地址有效性判断
		for index, fromAddress := range from {

			if IsValidForAddress([]byte(fromAddress)) == false || IsValidForAddress([]byte(to[index])) == false {

				fmt.Printf("Address:%s incalid", fromAddress)
				os.Exit(1)
			}
		}

		amount := Json2Array(*flagSendBlockAmount)

		cli.send(from, to, amount, nodeID, *flagSendBlockMine)
	}
	//对printchainCmd命令的解析
	if printchainCmd.Parsed() {

		cli.printchain(nodeID)
	}
	//创建区块链
	if createBlockchainCmd.Parsed() {

		if *flagCreateBlockchainAddress == "" {

			printUsage()
			os.Exit(1)
		}

		cli.creatBlockchain(*flagCreateBlockchainAddress, nodeID)
	}

	//查询余额
	if blanceBlockCmd.Parsed() {

		if *flagBlanceBlockAddress == "" {

			printUsage()
			os.Exit(1)
		}

		cli.getBlance(*flagBlanceBlockAddress, nodeID)
	}

	//创建钱包
	if createWalletCmd.Parsed() {

		cli.createWallet(nodeID)
	}

	//获取所有钱包地址
	if getAddressListCmd.Parsed() {

		cli.getAddressList(nodeID)
	}

	//UTXOSet测试
	if resetUTXOsetCmd.Parsed() {

		cli.ResetUTXOSet(nodeID)
	}

	//设置挖矿节点
	if startNodeCmd.Parsed() {

		cli.startNode(nodeID, *flagMiner)
	}
}
