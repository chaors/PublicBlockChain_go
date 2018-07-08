# goLang公链实战之boltdb数据库

我们知道，bitcoin客户端的区块信息是存储在LevelDB数据库中。我们既然要基于go开发公链，这里用到的数据库是基于go的[boltDB](https://github.com/boltdb)。

### 安装

使用go get
```
$ go get github.com/boltdb/boltd / ...
```

安装成功后，我们会在go目录下看到：

![boltdb安装目录](https://upload-images.jianshu.io/upload_images/830585-883a42d48e5e3518.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)

### 基本使用

##### 创建并打开数据库
注意：在这里gland直接运行，生成的my.db在main.go上层目录;命令行build在运行的话是当前目录！！！
```
//1.数据库创建
	//在这里gland直接运行，生成的my.db在main.go上层目录;命令行build在运行的话是当前目录！！！
	db, err := bolt.Open("chaorsBlock.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
```

在你打开之后，你有两种处理它的方式：读-写和只读操作，读-写方式开始于db.Update方法，常用于建表和表中插入新数据；只读操作开始于db.View方法，常用于表数据的查询。

##### 创建新表
```
//2.创建表
	err = db.Update(func(tx *bolt.Tx) error {
  
                //判断要创建的表是否存在
		b := tx.Bucket([]byte("MyBlocks"))
		if b == nil {
  
          		//创建叫"MyBucket"的表
			_, err := tx.CreateBucket([]byte("MyBlocks"))
			if err != nil {
                                //也可以在这里对表做插入操作
				log.Fatal(err)
			}
		}

		//一定要返回nil
		return nil
	})

	//更新数据库失败
	if err != nil {
		log.Fatal(err)
	}
```

##### 更新表内容
```
//3.更新表数据
	err = db.Update(func(tx *bolt.Tx) error {

		//取出叫"MyBucket"的表
		b := tx.Bucket([]byte("MyBlocks"))

		//往表里面存储数据
		if b != nil {
                        //插入的键值对数据类型必须是字节数组
			err := b.Put([]byte("l"), []byte("0x0000"))
			err := b.Put([]byte("ll"), []byte("0x0001"))
                        err := b.Put([]byte("lll"), []byte("0x0002"))
			if err != nil {
				log.Fatal(err)
			}
		}

		//一定要返回nil
		return nil
	})

	//更新数据库失败
	if err != nil {
		log.Fatal(err)
	}
```

##### 表查询
```
//4.查看表数据
	err = db.View(func(tx *bolt.Tx) error {

		//取出叫"MyBucket"的表
		b := tx.Bucket([]byte("MyBlocks"))

		//往表里面存储数据
		if b != nil {

			data := b.Get([]byte("l"))
			fmt.Printf("%s\n", data)
			data := b.Get([]byte("l"))
			fmt.Printf("%s\n", data)
		}

		//一定要返回nil
		return nil
	})

	//查询数据库失败
	if err != nil {
		log.Fatal(err)
	}
```

boltdb基本使用就先学到这，搭建公链用数据库存储区块大概也只用到这么多，以后具体涉及到boltdb其他知识再针对性学习就好。







