package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
)

func main() {
	c, err := tdx.Dial("124.71.187.122:7709")
	logs.PanicErr(err)

	resp, err := c.GetStockMinute(protocol.ExchangeSZ, "000001")
	logs.PrintErr(err)

	for _, v := range resp.List {
		logs.Debug(v)
	}

	<-c.Done()
}
