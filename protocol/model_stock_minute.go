package protocol

import (
	"errors"
)

type StockMinuteResp struct {
	Count uint16
	List  []PriceLevel
}

type stockMinute struct{}

func (this *stockMinute) Frame(exchange Exchange, code string) (*Frame, error) {
	if len(code) != 6 {
		return nil, errors.New("股票代码长度错误")
	}
	codeBs := []byte(code)
	codeBs = append(codeBs, 0x0, 0x0, 0x0, 0x0)
	return &Frame{
		Control: Control01,
		Type:    TypeStockMinute,
		Data:    append([]byte{exchange.Uint8(), 0x0}, codeBs...),
	}, nil
}

func (this *stockMinute) Decode(bs []byte) (*StockMinuteResp, error) {

	if len(bs) < 6 {
		return nil, errors.New("数据长度不足")
	}

	resp := &StockMinuteResp{
		Count: Uint16(bs[:2]),
	}
	//2-6字节是啥?
	bs = bs[6:]
	price := Price(0)

	for i := uint16(0); i < resp.Count; i++ {
		bs, price = GetPrice(bs)
		var unknown Price
		bs, unknown = GetPrice(bs) //这个是什么
		_ = unknown
		//logs.Debug(price, unknown)
		var number int
		bs, number = CutInt(bs)
		resp.List = append(resp.List, PriceLevel{
			Price:  price,
			Number: number,
		})
	}

	return resp, nil
}
