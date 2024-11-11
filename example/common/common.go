package common

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
)

func Test(f func(c *tdx.Client)) {
	for _, v := range tdx.Hosts {
		c, err := tdx.Dial(v, tdx.WithDebug())
		if err != nil {
			logs.PrintErr(err)
			continue
		}
		f(c)
		<-c.Done()
		break
	}
}
