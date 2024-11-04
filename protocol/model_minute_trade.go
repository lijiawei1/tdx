package protocol

import (
	"errors"
	"fmt"
)

type MinuteTradeReq struct {
	Exchange Exchange
	Code     string
	Start    uint16
	Count    uint16
}

func (req MinuteTradeReq) Check() error {
	if len(req.Code) != 6 {
		return errors.New("股票代码长度错误")
	}
	if req.Count > 1800 {
		return errors.New("数量不能超过1800")
	}
	return nil
}

type MinuteTradeResp struct {
	Count uint16
	List  []*MinuteTrade
}

// MinuteTrade 分时成交，todo 时间没有到秒，客户端上也没有,东方客户端能显示秒
type MinuteTrade struct {
	Time   string //时间
	Price  Price  //价格
	Volume int    //成交量
	Number int    //单数,历史数据改字段无效
	Status int    //0是买，1是卖，2无效（汇总出现）
}

func (this *MinuteTrade) String() string {
	return fmt.Sprintf("%s \t%-6s \t%-6s \t%-6d(手) \t%-4d(单) \t%-4s",
		this.Time, this.Price, this.Amount(), this.Volume, this.Number, this.StatusString())
}

// Amount 成交额
func (this *MinuteTrade) Amount() Price {
	return this.Price * Price(this.Volume) * 100
}

func (this *MinuteTrade) StatusString() string {
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
func (this *MinuteTrade) AvgVolume() float64 {
	return float64(this.Volume) / float64(this.Number)
}

// AvgPrice 平均每单成交金额
func (this *MinuteTrade) AvgPrice() Price {
	return Price(this.AvgVolume() * float64(this.Price) * 100)
}

// IsBuy 是否是买单
func (this *MinuteTrade) IsBuy() bool {
	return this.Status == 0
}

// IsSell 是否是卖单
func (this *MinuteTrade) IsSell() bool {
	return this.Status == 1
}

type minuteTrade struct{}

func (minuteTrade) Frame(req MinuteTradeReq) (*Frame, error) {
	if err := req.Check(); err != nil {
		return nil, err
	}
	codeBs := []byte(req.Code)
	codeBs = append(codeBs, Bytes(req.Start)...)
	codeBs = append(codeBs, Bytes(req.Count)...)
	return &Frame{
		Control: Control01,
		Type:    TypeMinuteTrade,
		Data:    append([]byte{req.Exchange.Uint8(), 0x0}, codeBs...),
	}, nil
}

func (minuteTrade) Decode(bs []byte, code string) (*MinuteTradeResp, error) {

	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	resp := &MinuteTradeResp{
		Count: Uint16(bs[:2]),
	}

	bs = bs[2:]

	lastPrice := Price(0)
	for i := uint16(0); i < resp.Count; i++ {
		mt := &MinuteTrade{
			Time: GetHourMinute([2]byte(bs[:2])),
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
