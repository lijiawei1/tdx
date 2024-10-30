package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
)

func main() {
	c, err := tdx.Dial("122.51.120.217:7709")
	logs.PanicErr(err)
	<-c.Done()
}
