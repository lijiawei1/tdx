### 拉取通达信的数据

1. 参考 https://github.com/bensema/gotdx
2. 参考 https://github.com/mootdx/mootdx

### 开发进度

* 基本信息(5档报价)
  ![](docs/plan20241025.png)
* 股票列表
  ![](docs/plan20241028-1.png)
* 分时成交
  ![](docs/plan20241028-2.png)
* K线
  ![](docs/plan20241029.png)

### 数据校对

* 日K线校对
  ![](docs/check_kline.png)
  ![](docs/check_kline_right.png)

* 校对分时成交
  ![](docs/check_trade.png)

### 如何使用

```go
package main

import (
	"fmt"
	"github.com/injoyai/tdx"
)

func main() {
	//连接服务器,开启日志,开启断连重试
	c, err := tdx.Dial("124.71.187.122:7709", tdx.WithDebug(), tdx.WithRedial())
	if err != nil {
		panic(err)
	}
	resp, err := c.GetQuote("sz000001", "sh600008")
	if err != nil {
		panic(err)
	}

	for _, v := range resp {
		fmt.Printf("%#v\n", v)
	}
	<-c.Done()
}

```

#### IP地址

| IP              | 检查时间       |
|-----------------|------------|
| 124.71.187.122  | 2024-10-30 |
| 122.51.120.217  | 2024-10-30 |
| 111.229.247.189 | 2024-11-04 |
| 124.70.199.56   | 2024-11-04 |
| 124.70.176.52   | 2024-11-04 |
| 124.70.133.119  | 2024-11-04 |
| 123.60.84.66    | 2024-11-04 |
| 123.60.73.44    | 2024-11-04 |
| 123.60.70.228   | 2024-11-04 |
| 123.60.186.45   | 2024-11-04 |
| 123.40.164.122  | 2024-11-04 |
| 123.249.15.60   | 2024-11-04 |
| 122.51.232.182  | 2024-11-04 |
| 121.36.81.195   | 2024-11-04 |
| 121.36.54.217   | 2024-11-04 |
| 121.36.225.169  | 2024-11-04 |
| 120.46.186.223  | 2024-11-04 |
| 119.97.185.59   | 2024-11-04 |
| 118.25.98.114   | 2024-11-04 |
| 116.205.183.150 | 2024-11-04 |
| 116.205.171.132 | 2024-11-04 |
| 116.205.163.254 | 2024-11-04 |
| 111.230.186.52  | 2024-11-04 |
| 110.41.4.4      | 2024-11-04 |
| 110.41.2.72     | 2024-11-04 |
| 110.41.154.219  | 2024-11-04 |
| 110.41.147.114  | 2024-11-04 |




