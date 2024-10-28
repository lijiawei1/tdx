package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
)

func main() {
	c, err := tdx.Dial("124.71.187.122:7709")
	logs.PanicErr(err)

	resp, err := c.GetStockList(protocol.ExchangeSH, 255)
	logs.PanicErr(err)

	for _, v := range resp.List {
		logs.Debug(v)
	}
	logs.Debug("总数:", resp.Count)

	select {}
}
