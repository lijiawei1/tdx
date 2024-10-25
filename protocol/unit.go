package protocol

import (
	"bytes"
	"fmt"
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

func FloatUnit(f float64) (float64, string) {
	m := []string{"万", "亿"}
	unit := ""
	for i := 0; f > 1e4 && i < len(m); f /= 1e4 {
		unit = m[i]
	}
	return f, unit
}

func FloatUnitString(f float64) string {
	m := []string{"万", "亿"}
	unit := ""
	for i := 0; f > 1e4 && i < len(m); f /= 1e4 {
		unit = m[i]
	}
	if unit == "" {
		return conv.String(f)
	}
	return fmt.Sprintf("%0.2f%s", f, unit)
}

func IntUnitString(n int) string {
	return FloatUnitString(float64(n))
}
