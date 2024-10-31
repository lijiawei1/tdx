package protocol

import (
	"errors"
	"fmt"
	"github.com/injoyai/conv"
)

// HistoryMinuteTradeAllReq 获取指定日期全部数据的请求参数
type HistoryMinuteTradeAllReq struct {
	Date     string //20241030
	Exchange Exchange
	Code     string
}

// HistoryMinuteTradeReq 获取指定日期分页数据的请求参数
type HistoryMinuteTradeReq struct {
	Date     string //20241030
	Exchange Exchange
	Code     string
	Start    uint16
	Count    uint16
}

func (req HistoryMinuteTradeReq) Check() error {
	if req.Count > 2000 {
		return errors.New("数量不能超过2000")
	}
	return nil
}

// HistoryMinuteTradeResp 历史分时交易比实时少了单量
type HistoryMinuteTradeResp struct {
	Count uint16
	List  []*HistoryMinuteTrade
}

type HistoryMinuteTrade struct {
	Time   string //时间
	Price  Price  //价格
	Volume int    //成交量
	Status int    //0是买，1是卖，2无效（汇总出现）
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

func (historyMinuteTrade) Frame(req HistoryMinuteTradeReq) (*Frame, error) {
	if err := req.Check(); err != nil {
		return nil, err
	}
	date := conv.Uint32(req.Date) //req.Time.Format("20060102"))
	dataBs := Bytes(date)
	dataBs = append(dataBs, req.Exchange.Uint8(), 0x0)
	dataBs = append(dataBs, []byte(req.Code)...)
	dataBs = append(dataBs, Bytes(req.Start)...)
	dataBs = append(dataBs, Bytes(req.Count)...)
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
		mt.Price = lastPrice / basePrice(code)
		bs, mt.Volume = CutInt(bs)
		bs, mt.Status = CutInt(bs)
		bs, _ = CutInt(bs) //这个得到的是0，不知道是啥
		resp.List = append(resp.List, mt)
	}

	return resp, nil
}
