package protocol

import "errors"

type StockCountResp struct {
	Count uint16
}

type stockCount struct{}

// Frame 0c0200000001080008004e04000075c73301
func (this *stockCount) Frame(exchange Exchange) *Frame {
	return &Frame{
		Control: Control01,
		Type:    TypeStockCount,
		Data:    []byte{exchange.Uint8(), 0x0, 0x75, 0xc7, 0x33, 0x01}, //后面的4字节不知道啥意思
	}
}

func (this *stockCount) Decode(bs []byte) (*StockCountResp, error) {
	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}
	return &StockCountResp{Count: Uint16(bs)}, nil
}
