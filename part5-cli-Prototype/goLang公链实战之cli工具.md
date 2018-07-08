# goLang公链实战之cli工具

goLang公链cli工具
[TOC]
我们在使用bitcoin客户端或者以太坊客户端时，都是可以在命令行直接执行挖矿，新增交易等操作。这些操作都有赖于cli工具，我们写的基于go的公链也会支持这个功能。所以，今天就学习实现简单的cli工具，以后再逐渐完善。

### flag:go下的命令行解析包

##### 常用的方法
```
//String defines a string flag with specified name, default value, and usage string.
//定义一个字符串的变量
type string string
func String(name string, value string, usage string) *string

//Int defines an int flag with specified name, default value, and usage string.
//定义一个int型变量
type int int
func Int(name string, value int, usage string) *int

//定义一个bool型变量
type bool bool
func Bool(name string, value bool, usage string) *bool

func Float64(name string, value float64, usage string) *float64

func Int64(name string, value int64, usage string) *int64

//在所有的变量参数被设置后需要进行一次解析才生效
func Parse()

//给指定的参数设置值
func Set(name, value string) error

//Flag结构体
type Flag struct {
        Name     string // name as it appears on command line
        Usage    string // help message
        Value    Value  // value as set
        DefValue string // default value (as text); for usage message
}
```

##### 简单示例
```
//定义一个字符串变量，并制定默认值以及使用方式
 flagStr := flag.String("printchain", "", "输出所有区块信息")

 //定义一个int型字符 默认值为6
 flagInt := flag.Int("number", 6, "输出一个数")

 //定义一个bool型变量 默认值为false
 flagBool := flag.Bool("open", false, "判断真假")

 //flag解析
 flag.Parse()
 
 fmt.Printf("%s\n", *flagStr)
 fmt.Printf("%d\n", *flagInt)
 fmt.Printf("%v\n", *flagBool)
```

##### 运行分析

![flag简单使用](https://upload-images.jianshu.io/upload_images/830585-7ca091bbddb28919.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

##### os.Args方法

go的os.Args方法可以获取用户在命令行输入的参数个数，跑个简单的🌰就知道它的用处了。

```
package main
import (
    "fmt"
    "os"
)

func main() {
    args := os.Args
    fmt.Printf("%v\n", args)
}
```
Run:

![os.Args](https://upload-images.jianshu.io/upload_images/830585-539e799375f1f7c2.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

不难发现，os.Args返回的是我们在命令行输入的参数数组，且顺序和输入顺序相同。

### cli工具

我们之前写好的添加区块，打印区块链都必须在.go文件中写好，然后编译运行。我们想要实现的是编译之后在命令行输入诸如以下命令再进行操作：
>//添加交易” chaors send 100btc to ww“到区块并进行挖矿
>./main -addBlock "chaors send 100btc to ww"

>//打印区块链的所有区块信息
>./main -printchain

以” addBlock“为例，我们必须这样做来达到上述目的：
__1.用flag设置addBlock命令__
__2. 用flag为addBlock命令设置参数__
__3.flag.Parse()解析参数__
__4.在解析中，调用区块链的新增区块方法新增区块__

##### CLI结构
```
type CLI struct {
 //添加cli命令行工具的类
 Blockchain *Blockchain
}
```

##### 当前所有方法的使用说明
```
//打印目前左右命令使用方法
func printUsage() {
 fmt.Println("Usage:")
 fmt.Println("\taddBlock -data DATA 交易信息   新增区块")
 fmt.Println("\tprintchain 打印所有区块信息")
}
```

##### 参数合法性验证

由于要在main后使用诸如addBlock命令调用相应方法，所有参数必须多余1个。当参数小于2个时，提示当前命令使用方法。

```
func isValidArgs() {
        
        //获取当前输入参数个数
        if len(os.Args) < 2 {
  printUsage()
  os.Exit(1)
 }
}
```

##### cli命令的真正调用

要想实现区块的新增，必须是命令行输入命令触发了相应的新增区块方法才行。这里这样定义：

```
//新增区块
func (cli *CLI) addBlock(data string) {
 cli.Blockchain.AddBlockToBlockchain(data)

}

//打印区块链
func (cli *CLI) printchain() {
 cli.Blockchain.Printchain()
}
```

##### 激活cli工具

```
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
```

##### cli测试

```
package main

import (

 "chaors.com/LearnGo/publicChaorsChain/part5-cli-Prototype/BLC"
)

func main() {
        
    //创建区块链
 blockchain := BLC.CreateBlockchainWithGensisBlock() 
    //创建命令行工具
 cli := BLC.CLI{blockchain}
    //激活cli
 cli.Run()
}
```
Run:

![cli_1](https://upload-images.jianshu.io/upload_images/830585-8262b8257ec08457.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

![cli_2](https://upload-images.jianshu.io/upload_images/830585-6fd57a160d86e8da.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

很显然，我们已经达到了预想的目的。当然，这还只是一个简陋的cli工具。以后，会随着go公链的开发进度而逐渐完善，今天就到这了。

