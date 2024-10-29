package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/example/common"
	"github.com/injoyai/tdx/protocol"
	"time"
)

func main() {
	common.Test(func(c *tdx.Client) {
		t := time.Date(2024, 10, 28, 0, 0, 0, 0, time.Local)
		resp, err := c.GetStockHistoryMinuteTrade(t, protocol.ExchangeSH, "000001", 0, 2000)
		logs.PanicErr(err)

		for _, v := range resp.List {
			logs.Debug(v)
		}

		logs.Debug("总数：", resp.Count)
	})
}
