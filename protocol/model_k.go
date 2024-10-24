package protocol

import (
	"fmt"
	"math"
)

// Price 价格，单位分
type Price int32

func (this Price) Float64() float64 {
	return float64(this) / 100
}

func (this Price) String() string {
	return fmt.Sprintf("%0.2f 元", this.Float64())
}

type PriceLevel struct {
	Price Price //价
	Vol   int   //量，是否是成交量？
}

// K k线图
type K struct {
	Last  Price //昨天收盘价
	Open  Price //今日开盘价
	High  Price //今日最高价
	Low   Price //今日最低价
	Close Price //今日收盘价
}

func (this K) String() string {
	return fmt.Sprintf("昨收:%0.2f, 今开:%0.2f, 最高:%0.2f, 最低:%0.2f, 今收:%0.2f", this.Last.Float64(), this.Open.Float64(), this.High.Float64(), this.Low.Float64(), this.Close.Float64())
}

func DecodeK(bs []byte) ([]byte, K) {
	k := K{}

	//当日收盘价
	bs, k.Close = GetPrice(bs)

	//前日收盘价
	bs, k.Last = GetPrice(bs)
	k.Last += k.Close

	//当日开盘价
	bs, k.Open = GetPrice(bs)
	k.Open += k.Close

	//当日最高价
	bs, k.High = GetPrice(bs)
	k.High += k.Close

	//当日最低价
	bs, k.Low = GetPrice(bs)
	k.Low += k.Close

	return bs, k
}

func GetPrice(bs []byte) ([]byte, Price) {
	for i := range bs {
		if bs[i]&0x80 == 0 {
			return bs[i+1:], getPrice(bs[:i+1])
		}
	}
	return bs, 0
}

/*
字节的第一位表示后续是否有数据（字节）
第一字节 的第二位表示正负 1负0正 有效数据为后6位
后续字节 的有效数据为后7位
最大长度未知
0x20说明有后续数据
*/
func getPrice(bs []byte) (data Price) {

	for i := range bs {
		switch i {
		case 0:
			//取后6位
			data += Price(int32(bs[0] & 0x3F))

		default:
			//取后7位
			data += Price(int32(bs[i]&0x7F) << uint8(6+(i-1)*7))

		}

		//判断是否有后续数据
		if bs[i]&0x80 == 0 {
			break
		}
	}

	//第一字节的第二位为1表示为负数
	if len(bs) > 0 && bs[0]&0x40 > 0 {
		data = -data
	}

	return
}

func CutInt(bs []byte) ([]byte, int) {
	for i := range bs {
		if bs[i]&0x80 == 0 {
			return bs[i+1:], getData(bs[:i+1])
		}
	}
	return bs, 0
}

func getData(bs []byte) (data int) {

	for i := range bs {
		switch i {
		case 0:
			//取后6位
			data += int(bs[0] & 0x3F)

		default:
			//取后7位
			data += int(bs[i]&0x7F) << uint8(6+(i-1)*7)

		}

		//判断是否有后续数据
		if bs[i]&0x80 == 0 {
			break
		}
	}

	//第一字节的第二位为1表示为负数
	if len(bs) > 0 && bs[0]&0x40 > 0 {
		data = -data
	}

	return
}

func getVolume(ivol uint32) (volume float64) {
	logpoint := ivol >> (8 * 3)
	//hheax := ivol >> (8 * 3)          // [3]
	hleax := (ivol >> (8 * 2)) & 0xff // [2]
	lheax := (ivol >> 8) & 0xff       //[1]
	lleax := ivol & 0xff              //[0]

	//dbl_1 := 1.0
	//dbl_2 := 2.0
	//dbl_128 := 128.0

	dwEcx := logpoint*2 - 0x7f
	dwEdx := logpoint*2 - 0x86
	dwEsi := logpoint*2 - 0x8e
	dwEax := logpoint*2 - 0x96
	tmpEax := dwEcx
	if dwEcx < 0 {
		tmpEax = -dwEcx
	} else {
		tmpEax = dwEcx
	}

	dbl_xmm6 := 0.0
	dbl_xmm6 = math.Pow(2.0, float64(tmpEax))
	if dwEcx < 0 {
		dbl_xmm6 = 1.0 / dbl_xmm6
	}

	dbl_xmm4 := 0.0
	dbl_xmm0 := 0.0

	if hleax > 0x80 {
		tmpdbl_xmm3 := 0.0
		//tmpdbl_xmm1 := 0.0
		dwtmpeax := dwEdx + 1
		tmpdbl_xmm3 = math.Pow(2.0, float64(dwtmpeax))
		dbl_xmm0 = math.Pow(2.0, float64(dwEdx)) * 128.0
		dbl_xmm0 += float64(hleax&0x7f) * tmpdbl_xmm3
		dbl_xmm4 = dbl_xmm0
	} else {
		if dwEdx >= 0 {
			dbl_xmm0 = math.Pow(2.0, float64(dwEdx)) * float64(hleax)
		} else {
			dbl_xmm0 = (1 / math.Pow(2.0, float64(dwEdx))) * float64(hleax)
		}
		dbl_xmm4 = dbl_xmm0
	}

	dbl_xmm3 := math.Pow(2.0, float64(dwEsi)) * float64(lheax)
	dbl_xmm1 := math.Pow(2.0, float64(dwEax)) * float64(lleax)
	if (hleax & 0x80) > 0 {
		dbl_xmm3 *= 2.0
		dbl_xmm1 *= 2.0
	}
	volume = dbl_xmm6 + dbl_xmm4 + dbl_xmm3 + dbl_xmm1
	return
}
