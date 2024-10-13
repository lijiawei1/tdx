package protocol

import (
	"encoding/hex"
	"testing"
)

func TestFrame_Bytes(t *testing.T) {
	f := Frame{
		MsgID:   1,
		Control: 1,
		Type:    Connect,
		Data:    []byte{0x01},
	}
	hex := f.Bytes().HEX()
	t.Log(hex)
	if hex != "0c0100000001030003000d0001" {
		t.Error("编码错误")
	}
}

func TestBytes(t *testing.T) {
	t.Log(hex.EncodeToString(Bytes(uint32(1))))
	t.Log(hex.EncodeToString(Bytes(uint16(0x0d00))))
}
