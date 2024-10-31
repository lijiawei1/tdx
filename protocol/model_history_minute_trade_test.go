package protocol

import (
	"testing"
	"time"
)

func Test_stockHistoryMinuteTrade_Frame(t *testing.T) {
	// 预期 0c 02000000 00 1200 1200 b50f 84da3401 0000 30303030303100006400
	//     0c000000000112001200b50f84da3401000030303030303100006400
	f, err := MStockHistoryMinuteTrade.Frame(StockHistoryMinuteTradeReq{
		Time:     time.Date(2024, 10, 28, 0, 0, 0, 0, time.Local),
		Exchange: ExchangeSZ,
		Code:     "000001",
		Start:    0,
		Count:    100,
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(f.Bytes().HEX())
}
