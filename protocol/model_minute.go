package protocol

import (
	"errors"
)

type MinuteResp struct {
	Count uint16
	List  []PriceNumber
}

type PriceNumber struct {
	Price  Price
	Number int
}

type minute struct{}

func (this *minute) Frame(code string) (*Frame, error) {
	exchange, number, err := DecodeCode(code)
	if err != nil {
		return nil, err
	}
	codeBs := []byte(number)
	codeBs = append(codeBs, 0x0, 0x0, 0x0, 0x0)
	return &Frame{
		Control: Control01,
		Type:    TypeMinute,
		Data:    append([]byte{exchange.Uint8(), 0x0}, codeBs...),
	}, nil
}

func (this *minute) Decode(bs []byte) (*MinuteResp, error) {

	if len(bs) < 6 {
		return nil, errors.New("数据长度不足")
	}

	resp := &MinuteResp{
		Count: Uint16(bs[:2]),
	}
	//2-6字节是啥?
	bs = bs[6:]
	price := Price(0)

	for i := uint16(0); i < resp.Count; i++ {
		bs, price = GetPrice(bs)
		bs, _ = CutInt(bs) //这个是什么
		var number int
		bs, number = CutInt(bs)
		resp.List = append(resp.List, PriceNumber{
			Price:  price,
			Number: number,
		})
	}

	return resp, nil
}
