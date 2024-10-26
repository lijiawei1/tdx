package tdx

import (
	"github.com/injoyai/base/maps/wait/v2"
	"github.com/injoyai/conv"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx/protocol"
	"time"
)

// Dial 与服务器建立连接
func Dial(addr string, op ...client.Option) (cli *Client, err error) {

	cli = &Client{
		w: wait.New(time.Second * 2),
	}

	cli.c, err = dial.TCP(addr, func(c *client.Client) {
		c.Logger.WithHEX() //以HEX显示
		c.SetOption(op...) //自定义选项
		//c.AllReader = ios.NewAllReader(c.Reader.(io.Reader), protocol.ReadFrom) //分包
		c.Event.OnReadFrom = protocol.ReadFrom         //分包
		c.Event.OnDealMessage = cli.handlerDealMessage //处理分包数据

		logs.Debug("option")
	})
	if err != nil {
		return nil, err
	}

	logs.Debug("run")
	logs.Debugf("%#v\n", cli.c.Event.OnReadFrom)
	go cli.c.Run()

	err = cli.connect()
	if err != nil {
		cli.c.Close()
		return nil, err
	}

	return cli, err
}

type Client struct {
	c     *client.Client
	w     *wait.Entity
	msgID uint32
}

// handlerDealMessage 处理服务器响应的数据
func (this *Client) handlerDealMessage(c *client.Client, msg ios.Acker) {

	f, err := protocol.Decode(msg.Payload())
	if err != nil {
		logs.Err(err)
		return
	}

	var resp any
	switch f.Type {
	case protocol.TypeSecurityQuote:
		resp = protocol.MSecurityQuote.Decode(f.Data)

	case protocol.TypeSecurityList:
		resp, err = protocol.MSecurityList.Decode(f.Data)

	}

	if err != nil {
		logs.Err(err)
		return
	}

	logs.Debug(resp)
	this.w.Done(conv.String(f.MsgID), resp)

}

func (this *Client) SendFrame(f *protocol.Frame) (any, error) {
	this.msgID++
	f.MsgID = this.msgID
	if _, err := this.c.Write(f.Bytes()); err != nil {
		return nil, err
	}
	return this.w.Wait(conv.String(this.msgID))
}

// Send 向服务发送数据，并等待响应数据
func (this *Client) Send(bs []byte) (any, error) {
	if _, err := this.c.Write(bs); err != nil {
		return nil, err
	}
	return this.w.Wait(conv.String(this.msgID))
}

// Write 实现io.Writer,向服务器写入数据
func (this *Client) Write(bs []byte) (int, error) {
	return this.c.Write(bs)
}

func (this *Client) Close() error {
	return this.c.Close()
}

func (this *Client) connect() error {
	f := protocol.MConnect.Frame()
	_, err := this.Write(f.Bytes())
	return err
}

// GetSecurityList 获取市场内指定范围内的所有证券代码
func (this *Client) GetSecurityList(exchange protocol.Exchange, starts ...uint16) (*protocol.SecurityListResp, error) {
	f := protocol.MSecurityList.Frame(exchange, starts...)
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.SecurityListResp), nil

}

// GetSecurityQuotes 获取盘口五档报价
func (this *Client) GetSecurityQuotes(m map[protocol.Exchange]string) (protocol.SecurityQuotesResp, error) {
	f, err := protocol.MSecurityQuote.Frame(m)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(protocol.SecurityQuotesResp), nil
}
