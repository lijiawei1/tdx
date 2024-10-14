package protocol

import (
	"bytes"
	bytes2 "github.com/injoyai/base/bytes"
	"github.com/injoyai/conv"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
)

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
