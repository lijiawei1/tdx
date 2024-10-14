package protocol

import "errors"

type ConnectResp struct {
	Info string
}

func DecodeConnect(bs []byte) (*ConnectResp, error) {
	if len(bs) < 68 {
		return nil, errors.New("数据长度不足")
	}
	//前68字节暂时还不知道是什么
	return &ConnectResp{Info: string(UTF8ToGBK(bs[68:]))}, nil
}

type SecurityListResp struct {
	Count uint16
	List  []*Security
}

type Security struct {
	Code         string
	VolUnit      uint16
	DecimalPoint int8
	Name         string
	PreClose     float64
}

func DecodeSecurityList(bs []byte) (*SecurityListResp, error) {

	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	count := Uint16(bs[:2])

	_ = count

	return nil, nil

}

func NewConnect() *Frame {
	return &Frame{
		Control: Control,
		Type:    Connect,
		Data:    []byte{0x01},
	}
}

func NewSecurityQuotes(m map[Exchange]string) (*Frame, error) {
	f := &Frame{
		Control: Control,
		Type:    SecurityQuote,
		Data:    []byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}

	payload := Bytes(uint16(len(m)))
	for k, v := range m {
		if len(v) != 6 {
			return nil, errors.New("股票代码长度错误")
		}
		payload = append(payload, k.Uint8())
		payload = append(payload, v...)
	}
	f.Data = append(f.Data, payload...)

	return f, nil
}
