package protocol

import (
	"testing"
	"time"
)

func Test_stockHistoryMinuteTrade_Frame(t *testing.T) {
	// 预期 0c 02000000 00 1200 1200 b50f 84da3401 0000 30303030303100006400
	//     0c000000000112001200b50f84da3401000030303030303100006400
	ti := time.Date(2024, 10, 28, 0, 0, 0, 0, time.Local)
	f, err := MStockHistoryMinuteTrade.Frame(ti, ExchangeSH, "000001", 0, 100)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(f.Bytes().HEX())
}
