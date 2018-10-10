## 操作说明


### 编译可执行文件
项目cmd 目录执行go build


### 使用可执行文件生成新账户
./cmd -c account lzhx_ createwallet     (lzhx_ 表示钱包后缀)

Your new address: 13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De

### 使用可执行文件查看钱包地址列表
./cmd -c account lzhx_ listaddresses  

13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De 

1LcubHwTs7AuHGxoPdwttbcbJkrrCwWUYX 

1zmkxjXmisf4kvQJCUnhKRYMbpsXBdPYf 

1EaKYX3U3GGjM4NoNXNVhMk7KQ1UKnTXgz 

1Pz7MTSoESDmnQDMaeLijG4qCbsPdvrus6 

### 启动链并连

接对端节点
./cmd -c chain -s lzhx_ -l 8080 -a 13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De -datadir /Users/username/goworkspace/src/Course/cmd

-s 表示 钱包后缀

-a 表示钱包地址

-datadir 表示数据存储目录，该参数为可选参数(该参数存在时，链数据将从指定目录存储和读取)


启动当前节点后日志输出一下信息

local http server listening on 127.0.0.1:8081
2018/09/21 15:06:20 I am /ip4/127.0.0.1/tcp/8080/ipfs/QmdhJPDZaLPCFjZMsuLfVtzZMNaZMPp6wT85gYdRnVcppj
2018/09/21 15:06:20 Now run "go run main.go -c chain -l 8082 -d /ip4/127.0.0.1/tcp/8080/ipfs/QmdhJPDZaLPCFjZMsuLfVtzZMNaZMPp6wT85gYdRnVcppj" on a different terminal

在另一个terminal中启动对端节点（加入-s参数表示该节点的钱包后缀名）

./cmd -s lzhx_ -c chain -l 8082 -d /ip4/127.0.0.1/tcp/8080/ipfs/QmdhJPDZaLPCFjZMsuLfVtzZMNaZMPp6wT85gYdRnVcppj

两节点正常连接后可在terminal中输入任意数字，该操作将产生新的块并同步块信息到对端节点

### 查看链状态、发送交易及交易打包

1.通过get形式的http请求查看链信息
e.g

path: http://127.0.0.1: &lt; port &gt;

return:
```json
   {
     "Blocks": [
       {
         "index": 0,
         "timestamp": "2018-09-20 17:51:25.453600292 +0800 CST m=+0.003102402",
         "result": 0,
         "hash": "f1534392279bddbf9d43dde8701cb5be14b82f76ec6607bf8d6ad557f60f304e",
         "prevhash": "",
         "proof": 100,
         "transactions": {},
         "accounts": {
           "13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De": 10000
         }
       },
       {
         "index": 1,
         "timestamp": "2018-09-20 17:51:27.893924841 +0800 CST m=+2.443408915",
         "result": 1,
         "hash": "1a32963946ab93f40d9cc9706503978c2854c86906fe691c4ebac989307a0671",
         "prevhash": "f1534392279bddbf9d43dde8701cb5be14b82f76ec6607bf8d6ad557f60f304e",
         "proof": 0,
         "transactions": null,
         "accounts": {}
       }
     ],
     "TxPool": {
       "AllTx": []
     }
   }
```


2.通过post形式的http接口发送交易到链上
e.g

path:   http://127.0.0.1: &lt; port &gt; /txpool

param:

```json
   {
    "From": "13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De",
    "To": "17eeNAJcUWECkHLDgGcXwZPKrYteNLq2hm",
    "Value": 100,
    "Data": "message"
}
```

return:
```json
    {
      "amount": 1,
      "recipient": "17eeNAJcUWECkHLDgGcXwZPKrYteNLq2hm",
      "sender": "13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De",
      "data": "bWVzc2FnZQ=="
    }
```



3.通过post形式的http接口发送信息产生新块
e.g

path:   http://127.0.0.1: &lt; port &gt; /block

param:

```json
    {"Msg": 123}
```

return:
```json
    {
      "index": 2,
      "timestamp": "2018-09-20 18:03:23.460148402 +0800 CST m=+24.501698347",
      "result": 123,
      "hash": "0ee7933883ae99f99fdc964042008426240066408ef8f0598e780a8158202f68",
      "prevhash": "e792220c169142a4561b7320005716a636a27b25bb2cb03c409a20ef64037d53",
      "proof": 0,
      "transactions": [
        {
          "amount": 1,
          "recipient": "17eeNAJcUWECkHLDgGcXwZPKrYteNLq2hm",
          "sender": "13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De8",
          "data": "bWVzc2FnZQ=="
        }
      ],
      "accounts": {
        "13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De": 9999,
        "17eeNAJcUWECkHLDgGcXwZPKrYteNLq2hm": 1
      }
    }
```


4.通过post形式的http接口查看账户余额
e.g

path:   http://127.0.0.1: &lt; port &gt; /getbalance

param:

```json
    {"Address": "17eeNAJcUWECkHLDgGcXwZPKrYteNLq2hm"}
```

return:
```
    1
```
