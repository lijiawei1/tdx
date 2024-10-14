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
