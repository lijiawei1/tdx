package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
)

func main() {
	c, err := tdx.Dial("124.71.187.122:7709")
	logs.PanicErr(err)

	resp, err := c.GetCode(protocol.ExchangeSH, 369)
	logs.PanicErr(err)

	for i, v := range resp.List {
		logs.Debug(i, v)
	}
	logs.Debug("总数:", resp.Count)

	select {}
}
