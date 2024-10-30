package protocol

import (
	"errors"
	"fmt"
	"github.com/injoyai/base/g"
	"time"
)

type StockKlineReq struct {
	Exchange Exchange
	Code     string
	Start    uint16
	Count    uint16
}

func (this *StockKlineReq) Bytes(Type TypeKline) (g.Bytes, error) {
	if len(this.Code) != 6 {
		return nil, errors.New("股票代码长度错误")
	}
	data := []byte{this.Exchange.Uint8(), 0x0}
	data = append(data, []byte(this.Code)...) //这里怎么是正序了？
	data = append(data, Bytes(Type.Uint16())...)
	data = append(data, 0x01, 0x0)
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
	Open   Price     //开盘价
	High   Price     //最高价
	Low    Price     //最低价
	Close  Price     //收盘价,如果是当天,则是最新价/实时价
	Volume float64   //成交量
	Amount float64   //成交额
	Time   time.Time //时间
}

func (this *StockKline) String() string {
	return fmt.Sprintf("%s 开盘价：%s 最高价：%s 最低价：%s 收盘价：%s 成交量：%s 成交额：%s",
		this.Time.Format("2006-01-02 15:04:05"),
		this.Open, this.High, this.Low, this.Close,
		FloatUnitString(this.Volume), FloatUnitString(this.Amount),
	)
}

type stockKline struct{}

func (stockKline) Frame(Type TypeKline, req *StockKlineReq) (*Frame, error) {
	bs, err := req.Bytes(Type)
	if err != nil {
		return nil, err
	}
	return &Frame{
		Control: Control01,
		Type:    TypeStockKline,
		Data:    bs,
	}, nil
}

func (stockKline) Decode(bs []byte, Type TypeKline) (*StockKlineResp, error) {

	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	resp := &StockKlineResp{
		Count: Uint16(bs[:2]),
	}

	bs = bs[2:]

	var last Price

	for i := uint16(0); i < resp.Count; i++ {
		k := &StockKline{
			Time: GetTime([4]byte(bs[:4]), Type),
		}

		var open Price
		bs, open = GetPrice(bs[4:])
		var _close Price
		bs, _close = GetPrice(bs)
		var high Price
		bs, high = GetPrice(bs)
		var low Price
		bs, low = GetPrice(bs)

		k.Open = (open + last) / 10
		k.Close = (open + last + _close) / 10
		k.High = (open + last + high) / 10
		k.Low = (open + last + low) / 10

		last = last + open + _close

		k.Volume = getVolume(Uint32(bs[:4]))
		k.Amount = getVolume(Uint32(bs[4:8]))

		bs = bs[8:]
		resp.List = append(resp.List, k)
	}

	return resp, nil
}
