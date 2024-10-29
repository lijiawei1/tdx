package protocol

import (
	"errors"
	"github.com/injoyai/conv"
	"time"
)

// StockHistoryMinuteTradeResp 历史分时交易比实时少了单量
type StockHistoryMinuteTradeResp struct {
	Count uint16
	List  []*StockMinuteTrade
}

type StockHistoryMinuteTrade struct {
	Time   string //时间
	Price  Price  //价格
	Volume int    //成交量
	Status int    //0是买，1是卖，2无效（汇总出现）
}

type stockHistoryMinuteTrade struct{}

func (stockHistoryMinuteTrade) Frame(t time.Time, exchange Exchange, code string, start, count uint16) (*Frame, error) {
	date := conv.Uint32(t.Format("20060102"))
	dataBs := Bytes(date)
	dataBs = append(dataBs, exchange.Uint8(), 0x0)
	dataBs = append(dataBs, []byte(code)...)
	dataBs = append(dataBs, Bytes(start)...)
	dataBs = append(dataBs, Bytes(count)...)
	return &Frame{
		Control: Control01,
		Type:    TypeStockHistoryMinuteTrade,
		Data:    dataBs,
	}, nil
}

func (stockHistoryMinuteTrade) Decode(bs []byte, code string) (*StockHistoryMinuteTradeResp, error) {
	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	resp := &StockHistoryMinuteTradeResp{
		Count: Uint16(bs[:2]),
	}

	//第2-6字节不知道是啥
	bs = bs[2+4:]

	lastPrice := Price(0)
	for i := uint16(0); i < resp.Count; i++ {
		mt := &StockMinuteTrade{
			Time: GetDate([2]byte(bs[:2])),
		}
		var sub Price
		bs, sub = GetPrice(bs[2:])
		lastPrice += sub
		mt.Price = lastPrice / basePrice(code)
		bs, mt.Volume = CutInt(bs)
		bs, mt.Status = CutInt(bs)
		bs, _ = CutInt(bs) //这个得到的是0，不知道是啥
		resp.List = append(resp.List, mt)
	}

	return resp, nil
}
