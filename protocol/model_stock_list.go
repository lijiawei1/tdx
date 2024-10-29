package protocol

import (
	"errors"
	"fmt"
)

type StockListResp struct {
	Count uint16
	List  []*Stock
}

type Stock struct {
	Name         string  //股票名称
	Code         string  //股票代码
	VolUnit      uint16  //未知
	DecimalPoint int8    //未知
	PreClose     float64 //未知
}

func (this *Stock) String() string {
	return fmt.Sprintf("%s(%s)", this.Code, this.Name)
}

type stockList struct{}

func (stockList) Frame(exchange Exchange, start uint16) *Frame {
	return &Frame{
		Control: Control01,
		Type:    TypeStockList,
		Data:    []byte{exchange.Uint8(), 0x0, uint8(start), uint8(start >> 8)},
	}
}

func (stockList) Decode(bs []byte) (*StockListResp, error) {

	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	resp := &StockListResp{
		Count: Uint16(bs[:2]),
	}
	bs = bs[2:]

	for i := uint16(0); i < resp.Count; i++ {
		sec := &Stock{
			Code:         string(bs[:6]),
			VolUnit:      Uint16(bs[6:8]),
			Name:         string(UTF8ToGBK(bs[8:16])),
			DecimalPoint: int8(bs[20]),
			PreClose:     getVolume(Uint32(bs[21:25])),
		}
		//logs.Debug(bs[25:29]) //26和28字节 好像是枚举(基本是44,45和34,35)
		bs = bs[29:]
		resp.List = append(resp.List, sec)
	}

	return resp, nil

}
