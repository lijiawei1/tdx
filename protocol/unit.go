package protocol

import (
	"bytes"
	bytes2 "github.com/injoyai/base/bytes"
	"github.com/injoyai/conv"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
)

func String(bs []byte) string {
	return string(bytes2.Reverse(bs))
}

func Bytes(n any) []byte {
	return bytes2.Reverse(conv.Bytes(n))
}

func Uint32(bs []byte) uint32 {
	return conv.Uint32(bytes2.Reverse(bs))
}

func Uint16(bs []byte) uint16 {
	return conv.Uint16(bytes2.Reverse(bs))
}

func UTF8ToGBK(text []byte) []byte {
	r := bytes.NewReader(text)
	decoder := transform.NewReader(r, simplifiedchinese.GBK.NewDecoder()) //GB18030
	content, _ := io.ReadAll(decoder)
	return bytes.ReplaceAll(content, []byte{0x00}, []byte{})
}

func getprice(b []byte, pos *int) int {
	/*
		    0x7f与常量做与运算实质是保留常量（转换为二进制形式）的后7位数，既取值区间为[0,127]
		    0x3f与常量做与运算实质是保留常量（转换为二进制形式）的后6位数，既取值区间为[0,63]

			0x80 1000 0000
			0x7f 0111 1111
			0x40  100 0000
			0x3f  011 1111
	*/
	posByte := 6
	bData := b[*pos]
	data := int(bData & 0x3f)
	bSign := false
	if (bData & 0x40) > 0 {
		bSign = true
	}

	if (bData & 0x80) > 0 {
		for {
			*pos += 1
			bData = b[*pos]
			data += (int(bData&0x7f) << posByte)

			posByte += 7

			if (bData & 0x80) <= 0 {
				break
			}
		}
	}
	*pos++

	if bSign {
		data = -data
	}
	return data
}
