package tdx

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx/protocol"
	"testing"
	"time"
)

var (
	c  *Client
	do func(f func(c *Client))
)

func init() {
	var err error
	c, err = Dial("124.71.187.122:7709")
	logs.PanicErr(err)
	do = func(f func(c *Client)) {
		f(c)
		<-c.Done()
	}
}

func TestClient_GetStockHistoryMinuteTrade(t *testing.T) {
	ti := time.Date(2024, 10, 28, 0, 0, 0, 0, time.Local)
	do(func(c *Client) {
		resp, err := c.GetStockHistoryMinuteTrade(ti, protocol.ExchangeSH, "000001", 0, 100)
		if err != nil {
			t.Error(err)
			return
		}
		for _, v := range resp.List {
			t.Log(v)
		}
		t.Log("总数：", resp.Count)
	})

}
