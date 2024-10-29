package protocol

import (
	"errors"
	"github.com/injoyai/base/g"
	"time"
)

type StockKlineReq struct {
	Exchange Exchange
	Code     string
	Type     uint16 //类型 1分 5分 等等
	Start    uint16
	Count    uint16
}

func (this *StockKlineReq) Bytes() (g.Bytes, error) {
	if len(this.Code) != 6 {
		return nil, errors.New("股票代码长度错误")
	}
	data := []byte{this.Exchange.Uint8(), 0x0}
	data = append(data, Bytes(this.Code)...)
	data = append(data, Bytes(this.Type)...)
	data = append(data, 0x0, 0x0)
	data = append(data, Bytes(this.Start)...)
	data = append(data, Bytes(this.Count)...)
	data = append(data, make([]byte, 10)...) //未知啥含义
	return data, nil
}

type StockKlineResp struct {
	Count uint16
	List  []*StockKline
}

type StockKline struct {
	Open   Price
	High   Price
	Low    Price
	Close  Price
	Volume int //成交量
	Number int //成交数
	Time   time.Time
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
}

type stockKline struct{}

func (stockKline) Frame(req *StockKlineReq) (*Frame, error) {
	bs, err := req.Bytes()
	if err != nil {
		return nil, err
	}
	return &Frame{
		Control: Control01,
		Type:    TypeStockKline,
		Data:    bs,
	}, nil
}

func (stockKline) Decode(bs []byte) (*StockKlineResp, error) {

	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	resp := &StockKlineResp{
		Count: Uint16(bs[:2]),
	}

	for i := uint16(0); i < resp.Count; i++ {

	}

	return nil, nil
}
