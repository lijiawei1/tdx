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

// String 字节先转小端,再转字符
func String(bs []byte) string {
	return string(bytes2.Reverse(bs))
}

// Bytes 任意类型转小端字节
func Bytes(n any) []byte {
	return bytes2.Reverse(conv.Bytes(n))
}

// Uint32 字节通过小端方式转为uint32
func Uint32(bs []byte) uint32 {
	return conv.Uint32(bytes2.Reverse(bs))
}

// Uint16 字节通过小端方式转为uint16
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

func GetTime(bs [2]byte) string {
	n := Uint16(bs[:])
	h := n / 60
	m := n % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func basePrice(code string) Price {
	if len(code) == 0 {
		return 1
	}
	switch code[:2] {
	case "60", "30", "68", "00":
		return 1
	default:
		return 10
	}
}
