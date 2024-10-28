package protocol

import (
	"errors"
	"fmt"
)

type StockMinuteTradeResp struct {
	Count uint16
	List  []*StockMinuteTrade
}

type StockMinuteTrade struct {
	Time   string //时间
	Price  Price  //价格
	Volume int    //成交量
	Number int    //单数
	Status int    //0是买，1是卖，2无效（汇总出现）
}

func (this *StockMinuteTrade) String() string {
	return fmt.Sprintf("%s \t%s \t%-6d(手) \t%-4d(单) \t%-4s", this.Time, this.Price, this.Volume, this.Number, this.StatusString())
}

func (this *StockMinuteTrade) StatusString() string {
	switch this.Status {
	case 0:
		return "买入"
	case 1:
		return "卖出"
	default:
		return ""
	}
}

// AvgVolume 平均每单成交量
func (this *StockMinuteTrade) AvgVolume() float64 {
	return float64(this.Volume) / float64(this.Number)
}

// AvgPrice 平均每单成交金额
func (this *StockMinuteTrade) AvgPrice() Price {
	return Price(this.AvgVolume() * float64(this.Price) * 100)
}

// IsBuy 是否是买单
func (this *StockMinuteTrade) IsBuy() bool {
	return this.Status == 0
}

// IsSell 是否是卖单
func (this *StockMinuteTrade) IsSell() bool {
	return this.Status == 1
}

type stockMinuteTrade struct{}

func (stockMinuteTrade) Frame(exchange Exchange, code string, start, count uint16) (*Frame, error) {
	if len(code) != 6 {
		return nil, errors.New("股票代码长度错误")
	}
	codeBs := []byte(code)
	codeBs = append(codeBs, Bytes(start)...)
	codeBs = append(codeBs, Bytes(count)...)
	return &Frame{
		Control: Control01,
		Type:    TypeStockMinuteTrade,
		Data:    append([]byte{exchange.Uint8(), 0x0}, codeBs...),
	}, nil
}

func (stockMinuteTrade) Decode(bs []byte, code string) (*StockMinuteTradeResp, error) {

	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	resp := &StockMinuteTradeResp{
		Count: Uint16(bs[:2]),
	}

	bs = bs[2:]

	lastPrice := Price(0)
	for i := uint16(0); i < resp.Count; i++ {
		mt := &StockMinuteTrade{
			Time: GetTime([2]byte(bs[:2])),
		}
		var sub Price
		bs, sub = GetPrice(bs[2:])
		lastPrice += sub
		mt.Price = lastPrice / basePrice(code)
		bs, mt.Volume = CutInt(bs)
		bs, mt.Number = CutInt(bs)
		bs, mt.Status = CutInt(bs)
		bs, _ = CutInt(bs) //这个得到的是0，不知道是啥
		resp.List = append(resp.List, mt)
	}

	return resp, nil
}
