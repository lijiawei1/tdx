### 拉取通达信的数据
### 参考 https://github.com/bensema/gotdx


### 开发进度
* 基本信息(5档报价)
![](docs/plan20241025.png)
* 股票列表
![](docs/plan20241028-1.png)
* 分时成交
![](docs/plan20241028-2.png)
* K线
![](docs/plan20241029.png)

### 如何使用

```go
package main

import (
	"fmt"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
)

func main() {
	//连接服务器,开启日志,开启断连重试
	c, err := tdx.Dial("124.71.187.122:7709", tdx.WithDebug(true), tdx.WithRedial(true))
	if err != nil {
		panic(err)
	}
	resp, err := c.GetStockQuotes(map[protocol.Exchange]string{
		protocol.ExchangeSH: "000001",
		protocol.ExchangeSZ: "600008",
	})
	if err != nil {
		panic(err)
	}

	for _, v := range resp {
		fmt.Printf("%#v\n", v)
	}
	<-c.Done()
}

```