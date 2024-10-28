package common

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
)

func Test(f func(c *tdx.Client)) {
	c, err := tdx.Dial("124.71.187.122:7709")
	logs.PanicErr(err)
	f(c)
	<-c.Done()
}
