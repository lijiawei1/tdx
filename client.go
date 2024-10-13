package tdx

import (
	"encoding/hex"
	"github.com/injoyai/base/maps/wait/v2"
	"github.com/injoyai/conv"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx/protocol"
	"time"
)

func Dial(addr string, op ...client.Option) (*Client, error) {
	c, err := dial.TCP(addr, func(c *client.Client) {
		c.Logger.WithHEX() //以HEX显示
		c.SetOption(op...) //自定义选项
		//c.Event.OnReadFrom = protocol.ReadFrom     //分包
		c.Event.OnDealMessage = handlerDealMessage //处理分包数据
	})
	if err != nil {
		return nil, err
	}

	go c.Run()

	cli := &Client{
		c: c,
		w: wait.New(time.Second * 2),
	}

	err = cli.Connect()

	return cli, err
}

// handlerDealMessage 处理服务器响应的数据
func handlerDealMessage(c *client.Client, msg ios.Acker) {

	f, err := protocol.Decode(msg.Payload())
	if err != nil {
		logs.Err(err)
		return
	}

	_ = f

}

type Client struct {
	c     *client.Client
	w     *wait.Entity
	msgID uint32
}

func (this *Client) SendFrame(f protocol.Frame) (any, error) {
	this.msgID++
	f.MsgID = this.msgID
	if _, err := this.c.Write(f.Bytes()); err != nil {
		return nil, err
	}
	return this.w.Wait(conv.String(this.msgID))
}

func (this *Client) Send(bs []byte) (any, error) {
	if _, err := this.c.Write(bs); err != nil {
		return nil, err
	}
	return this.w.Wait(conv.String(this.msgID))
}

func (this *Client) Write(bs []byte) (int, error) {
	return this.c.Write(bs)
}

func (this *Client) Close() error {
	return this.c.Close()
}

func (this *Client) Connect() error {
	f := protocol.Frame{
		Control: 0x01,
		Type:    protocol.Connect,
		Data:    []byte{0x01},
	}
	_, err := this.Write(f.Bytes())
	return err
}

// GetSecurityList 获取市场内指定范围内的所有证券代码
// 0c02000000011a001a003e05050000000000000002000030303030303101363030303038
func (this *Client) GetSecurityList() (*protocol.SecurityListResp, error) {

	f := protocol.Frame{
		Control: 0x01,
		Type:    protocol.Connect,
		Data:    nil,
	}
	_ = f

	bs, err := hex.DecodeString("0c02000000011a001a003e05050000000000000002000030303030303101363030303038")
	if err != nil {
		return nil, err
	}

	_, err = this.Write(bs)
	return nil, err

}
