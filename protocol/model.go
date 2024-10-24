package protocol

import (
	"errors"
	"fmt"
	"strings"
)

var (
	MConnect       = connect{}
	MSecurityQuote = securityQuote{}
	SecurityList   = securityList{}
)

type ConnectResp struct {
	Info string
}

type connect struct{}

func (connect) Frame() *Frame {
	return &Frame{
		Control: Control,
		Type:    TypeConnect,
		Data:    []byte{0x01},
	}
}

func (connect) Decode(bs []byte) (*ConnectResp, error) {
	if len(bs) < 68 {
		return nil, errors.New("数据长度不足")
	}
	//前68字节暂时还不知道是什么
	return &ConnectResp{Info: string(UTF8ToGBK(bs[68:]))}, nil
}

/*



 */

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

type securityList struct{}

func (securityList) Frame() *Frame {
	return &Frame{
		Control: 0x01,
		Type:    TypeConnect,
		Data:    nil,
	}
}

func (securityList) Decode(bs []byte) (*SecurityListResp, error) {

	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	count := Uint16(bs[:2])

	_ = count

	return nil, nil

}

/*



 */

type SecurityQuotesResp []*SecurityQuote

func (this SecurityQuotesResp) String() string {
	ls := []string(nil)
	for _, v := range this {
		ls = append(ls, v.String())
	}
	return strings.Join(ls, "\n")
}

type SecurityQuote struct {
	Exchange       Exchange // 市场
	Code           string   // 代码
	Active1        uint16   // 活跃度
	K              K        //k线
	ServerTime     string   // 时间
	ReversedBytes0 int      // 保留(时间 ServerTime)
	ReversedBytes1 int      // 保留
	TotalHand      int      // 总手（东财的盘口-总手）
	Intuition      int      // 现量（东财的盘口-现量）
	Amount         float64  // 金额（东财的盘口-金额）
	InsideDish     int      // 内盘（东财的盘口-外盘）（和东财对不上）
	OuterDisc      int      // 外盘（东财的盘口-外盘）（和东财对不上）

	ReversedBytes2 int // 保留
	ReversedBytes3 int // 保留
	BidLevels      [5]PriceLevel
	AskLevels      [5]PriceLevel
	Bid1           float64
	Ask1           float64
	BidVol1        int
	AskVol1        int
	Bid2           float64
	Ask2           float64
	BidVol2        int
	AskVol2        int
	Bid3           float64
	Ask3           float64
	BidVol3        int
	AskVol3        int
	Bid4           float64
	Ask4           float64
	BidVol4        int
	AskVol4        int
	Bid5           float64
	Ask5           float64
	BidVol5        int
	AskVol5        int
	ReversedBytes4 uint16  // 保留
	ReversedBytes5 int     // 保留
	ReversedBytes6 int     // 保留
	ReversedBytes7 int     // 保留
	ReversedBytes8 int     // 保留
	ReversedBytes9 uint16  // 保留
	Rate           float64 // 涨速
	Active2        uint16  // 活跃度
}

func (this *SecurityQuote) String() string {
	return this.K.String() + fmt.Sprintf(", 总量：%s, 现量：%s, 总金额：%s, 内盘：%s, 外盘：%s",
		IntUnitString(this.TotalHand), IntUnitString(this.Intuition), FloatUnitString(this.Amount),
		IntUnitString(this.InsideDish), IntUnitString(this.OuterDisc)) + "\n" +
		fmt.Sprintf("%#v\n", this)
}

type securityQuote struct{}

func (this securityQuote) Frame(m map[Exchange]string) (*Frame, error) {
	f := &Frame{
		Control: Control,
		Type:    TypeSecurityQuote,
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

/*
Decode
b1cb74000c02000000003e05af00af000136020000303030303031320bb2124c56105987e6d10cf212b78fa801ae01293dc54e8bd740acb8670086ca1e0001af36ba0c4102b467b6054203a68a0184094304891992114405862685108d0100000000e8ff320b0136303030303859098005464502468defd10cc005bed2668e05be15804d8ba12cb3b13a0083c3034100badc029d014201bc990384f70443029da503b7af074403a6e501b9db044504a6e2028dd5048d050000000000005909
*/
func (this securityQuote) Decode(bs []byte) SecurityQuotesResp {

	resp := SecurityQuotesResp{}

	//前2字节是什么?
	bs = bs[2:]

	number := Uint16(bs[:2])
	bs = bs[2:]

	for i := uint16(0); i < number; i++ {
		sec := &SecurityQuote{
			Exchange: Exchange(bs[0]),
			Code:     string(UTF8ToGBK(bs[1:7])),
			Active1:  Uint16(bs[7:9]),
		}
		bs, sec.K = DecodeK(bs[9:])
		bs, sec.ReversedBytes0 = CutInt(bs)
		sec.ServerTime = fmt.Sprintf("%d", sec.ReversedBytes0)
		bs, sec.ReversedBytes1 = CutInt(bs)
		bs, sec.TotalHand = CutInt(bs)
		bs, sec.Intuition = CutInt(bs)
		sec.Amount = getVolume(Uint32(bs[:4]))
		bs, sec.InsideDish = CutInt(bs[4:])
		bs, sec.OuterDisc = CutInt(bs)
		bs, sec.ReversedBytes2 = CutInt(bs)
		bs, sec.ReversedBytes3 = CutInt(bs)

		var p Price
		for i := 0; i < 5; i++ {
			bidele := PriceLevel{}
			offerele := PriceLevel{}

			bs, p = GetPrice(bs)
			bidele.Price = p + sec.K.Close
			bs, p = GetPrice(bs)
			offerele.Price = p + sec.K.Close

			bs, bidele.Vol = CutInt(bs)
			bs, offerele.Vol = CutInt(bs)

			sec.BidLevels[i] = bidele
			sec.AskLevels[i] = offerele
		}

		sec.ReversedBytes4 = Uint16(bs[:2])
		bs, sec.ReversedBytes5 = CutInt(bs[2:])
		bs, sec.ReversedBytes6 = CutInt(bs)
		bs, sec.ReversedBytes7 = CutInt(bs)
		bs, sec.ReversedBytes8 = CutInt(bs)
		sec.ReversedBytes9 = Uint16(bs[:2])

		sec.Rate = float64(sec.ReversedBytes9) / 100
		sec.Active2 = Uint16(bs[2:4])

		resp = append(resp, sec)
	}

	return resp
}
