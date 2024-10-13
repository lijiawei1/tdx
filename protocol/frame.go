package protocol

import (
	"errors"
	"github.com/injoyai/base/bytes"
	"github.com/injoyai/base/g"
	"github.com/injoyai/conv"
	"io"
)

const (
	// Prefix 固定帧头
	Prefix = 0x0c
)

type Message interface {
	Bytes() g.Bytes
}

/*
Frame 数据帧
0c 02189300 01 0300 0300 0d00 01
0c 00000000 00 0200 0200 1500
0c 01000000 01 0300 0300 0d00 01
0c 01000000 01 0300 0300 0d00 01
0c 02000000 01 1a00 1a00 3e05 050000000000000002000030303030303101363030303038

0c0100000001030003000d0001
*/
type Frame struct {
	MsgID   uint32 //消息ID
	Control uint8  //控制码，这个还不知道怎么定义
	Type    uint16 //请求类型，如建立连接，请求分时数据等
	Data    []byte //数据
}

func (this *Frame) Bytes() g.Bytes {
	length := uint16(len(this.Data) + 2)
	data := make([]byte, 12+len(this.Data))
	data[0] = Prefix
	copy(data[1:], Bytes(this.MsgID))
	data[5] = this.Control
	copy(data[6:], Bytes(length))
	copy(data[8:], Bytes(length))
	copy(data[10:], Bytes(this.Type))
	copy(data[12:], this.Data)
	return data
}

func Bytes(n any) []byte {
	return bytes.Reverse(conv.Bytes(n))
}

func Decode(bs []byte) (*Frame, error) {
	if len(bs) < 10 {
		return nil, errors.New("数据长度不足")
	}
	f := &Frame{}

	return f, nil
}

// ReadFrom 这里的r推荐传入*bufio.Reader
func ReadFrom(r io.Reader) ([]byte, error) {
	result := []byte(nil)
	b := make([]byte, 1)
	for {
		result = []byte(nil)

		n, err := r.Read(b)
		if err != nil {
			return nil, err
		}
		if n == 0 || b[0] != Prefix {
			continue
		}

		result = append(result, b[0])

		//读取9字节 消息ID+控制码+2个字节长度
		buf := make([]byte, 9)
		n, err = r.Read(buf)
		if err != nil {
			return nil, err
		}
		if n != 9 {
			continue
		}
		result = append(result, buf...)

		//获取后续字节长度
		length := uint16(result[9])<<8 + uint16(result[10])
		buf = make([]byte, length)
		n, err = r.Read(buf)
		if err != nil {
			return nil, err
		}
		if n != int(length) {
			continue
		}
		result = append(result, buf...)

		return result, nil
	}

}
