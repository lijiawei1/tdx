package protocol

import (
	"errors"
	"fmt"
	"github.com/injoyai/base/g"
	"time"
)

type KlineReq struct {
	Exchange Exchange
	Code     string
	Start    uint16
	Count    uint16
}

func (this *KlineReq) Bytes(Type uint8) (g.Bytes, error) {
	if this.Count > 800 {
		return nil, errors.New("单次数量不能超过800")
	}
	if len(this.Code) != 6 {
		return nil, errors.New("股票代码长度错误")
	}
	data := []byte{this.Exchange.Uint8(), 0x0}
	data = append(data, []byte(this.Code)...) //这里怎么是正序了？
	data = append(data, Type, 0x0)
	data = append(data, 0x01, 0x0)
	data = append(data, Bytes(this.Start)...)
	data = append(data, Bytes(this.Count)...)
	data = append(data, make([]byte, 10)...) //未知啥含义
	return data, nil
}

type KlineResp struct {
	Count uint16
	List  []*Kline
}

type Kline struct {
	Last   Price     //昨日收盘价,这个是列表的上一条数据的收盘价，如果没有上条数据，那么这个值为0
	Open   Price     //开盘价
	High   Price     //最高价
	Low    Price     //最低价
	Close  Price     //收盘价,如果是当天,则是最新价/实时价
	Volume float64   //成交量
	Amount float64   //成交额
	Time   time.Time //时间
}

func (this *Kline) String() string {
	return fmt.Sprintf("%s 昨收盘：%s 开盘价：%s 最高价：%s 最低价：%s 收盘价：%s 涨跌：%s 涨跌幅：%0.2f 成交量：%s 成交额：%s",
		this.Time.Format("2006-01-02 15:04:05"),
		this.Last, this.Open, this.High, this.Low, this.Close,
		this.RisePrice(), this.RiseRate(),
		FloatUnitString(this.Volume), FloatUnitString(this.Amount),
	)
}

// MaxDifference 最大差值，最高-最低
func (this *Kline) MaxDifference() Price {
	return this.High - this.Low
}

// RisePrice 涨跌金额,第一个数据不准，仅做参考
func (this *Kline) RisePrice() Price {
	if this.Last == 0 {
		//稍微数据准确点，没减去0这么夸张，还是不准的
		return this.Close - this.Open
	}
	return this.Close - this.Last

}

// RiseRate 涨跌比例/涨跌幅,第一个数据不准，仅做参考
func (this *Kline) RiseRate() float64 {
	return float64(this.RisePrice()) / float64(this.Open) * 100
}

type kline struct{}

func (kline) Frame(Type uint8, req KlineReq) (*Frame, error) {
	bs, err := req.Bytes(Type)
	if err != nil {
		return nil, err
	}
	return &Frame{
		Control: Control01,
		Type:    TypeKline,
		Data:    bs,
	}, nil
}

func (kline) Decode(bs []byte, Type uint8) (*KlineResp, error) {

	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	resp := &KlineResp{
		Count: Uint16(bs[:2]),
	}

	bs = bs[2:]

	var last Price //上条数据(昨天)的收盘价
	for i := uint16(0); i < resp.Count; i++ {
		k := &Kline{
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

		k.Last = last / 10
		k.Open = (open + last) / 10
		k.Close = (last + open + _close) / 10
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
