package protocol

import (
	"errors"
	"time"
)

type IndexResp struct {
	Count uint16
	List  []*IndexKline
}

type IndexKline struct {
	Last      Price     //昨日收盘价,这个是列表的上一条数据的收盘价，如果没有上条数据，那么这个值为0
	Open      Price     //开盘价
	High      Price     //最高价
	Low       Price     //最低价
	Close     Price     //收盘价,如果是当天,则是最新价/实时价
	Volume    int64     //成交量
	Amount    Price     //成交额
	Time      time.Time //时间
	UpCount   uint16    //
	DownCount uint16    //
}

type index struct{}

func (index) Frame() *Frame {
	return &Frame{
		Control: Control01,
		Type:    TypeIndex,
		Data:    make([]byte, 10),
	}
}

func (index) Decode(bs []byte, Type uint8) (*IndexResp, error) {

	if len(bs) < 2 {
		return nil, errors.New("数据长度不足")
	}

	resp := &IndexResp{
		Count: Uint16(bs[:2]),
	}

	bs = bs[2:]

	var last Price //上条数据(昨天)的收盘价
	for i := uint16(0); i < resp.Count; i++ {
		k := &IndexKline{
			Time: GetTime([4]byte(bs[:4]), Type),
		}

		var open Price
		bs, open = GetPrice(bs[4:])
		var _close Price
		bs, _close = GetPrice(bs)
		var high Price
		bs, high = GetPrice(bs)
		var low Price
		bs, low = GetPrice(bs)

		k.Last = last / 10
		k.Open = (open + last) / 10
		k.Close = (last + open + _close) / 10
		k.High = (open + last + high) / 10
		k.Low = (open + last + low) / 10
		last = last + open + _close

		k.Volume = int64(getVolume(Uint32(bs[:4])))
		switch Type {
		case TypeKlineMinute, TypeKline5Minute, TypeKlineMinute2, TypeKline15Minute, TypeKline30Minute, TypeKlineHour, TypeKlineDay2:
			k.Volume /= 100
		}
		k.Amount = Price(getVolume(Uint32(bs[4:8])) * 100) //从元转为分,并去除多余的小数

		resp.List = append(resp.List, k)
	}

	return resp, nil
}
