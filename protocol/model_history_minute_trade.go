package protocol

import (
	"errors"
	"fmt"
	"github.com/injoyai/conv"
)

// HistoryMinuteTradeResp 历史分时交易比实时少了单量
type HistoryMinuteTradeResp struct {
	Count uint16
	List  []*HistoryMinuteTrade
}

type HistoryMinuteTrade struct {
	Time   string //时间
	Price  Price  //价格
	Volume int    //成交量
	Status int    //0是买，1是卖，2无效（汇总出现）中途也可能出现2,例20241115(sz000001)的14:56
}

func (this *HistoryMinuteTrade) String() string {
	return fmt.Sprintf("%s \t%s \t%-6s \t%-6d(手) \t%-4s", this.Time, this.Price, this.Amount(), this.Volume, this.StatusString())
}

// Amount 成交额
func (this *HistoryMinuteTrade) Amount() Price {
	return this.Price * Price(this.Volume) * 100
}

func (this *HistoryMinuteTrade) StatusString() string {
	switch this.Status {
	case 0:
		return "买入"
	case 1:
		return "卖出"
	default:
		return ""
	}
}

type historyMinuteTrade struct{}

func (historyMinuteTrade) Frame(date, code string, start, count uint16) (*Frame, error) {
	exchange, number, err := DecodeCode(code)
	if err != nil {
		return nil, err
	}
	dataBs := Bytes(conv.Uint32(date)) //req.Time.Format("20060102"))
	dataBs = append(dataBs, exchange.Uint8(), 0x0)
	dataBs = append(dataBs, []byte(number)...)
	dataBs = append(dataBs, Bytes(start)...)
	dataBs = append(dataBs, Bytes(count)...)
	return &Frame{
		Control: Control01,
		Type:    TypeHistoryMinuteTrade,
		Data:    dataBs,
	}, nil
}

func (historyMinuteTrade) Decode(bs []byte, code string) (*HistoryMinuteTradeResp, error) {
	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	_, number, err := DecodeCode(code)
	if err != nil {
		return nil, err
	}

	resp := &HistoryMinuteTradeResp{
		Count: Uint16(bs[:2]),
	}

	//第2-6字节不知道是啥
	bs = bs[2+4:]

	lastPrice := Price(0)
	for i := uint16(0); i < resp.Count; i++ {
		mt := &HistoryMinuteTrade{
			Time: GetHourMinute([2]byte(bs[:2])),
		}
		var sub Price
		bs, sub = GetPrice(bs[2:])
		lastPrice += sub
		mt.Price = lastPrice / basePrice(number)
		bs, mt.Volume = CutInt(bs)
		bs, mt.Status = CutInt(bs)
		bs, _ = CutInt(bs) //这个得到的是0，不知道是啥
		resp.List = append(resp.List, mt)
	}

	return resp, nil
}
