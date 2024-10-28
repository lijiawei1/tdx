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
		c.Logger.WithHEX()                             //以HEX显示
		c.SetOption(op...)                             //自定义选项
		c.Event.OnReadFrom = protocol.ReadFrom         //分包
		c.Event.OnDealMessage = cli.handlerDealMessage //处理分包数据
		//无数据超时时间是60秒
		c.GoTimerWriter(30*time.Second, func(w ios.MoreWriter) error {
			bs := protocol.MHeart.Frame().Bytes()
			_, err := w.Write(bs)
			return err
		})
	})
	if err != nil {
		return nil, err
	}

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

// Done 连接关闭
func (this *Client) Done() <-chan struct{} {
	return this.c.Done()
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

	case protocol.TypeConnect:

	case protocol.TypeStockList:
		resp, err = protocol.MStockList.Decode(f.Data)

	case protocol.TypeStockQuote:
		resp = protocol.MStockQuote.Decode(f.Data)

	case protocol.TypeStockMinute:
		resp, err = protocol.MStockMinute.Decode(f.Data)

	}

	if err != nil {
		logs.Err(err)
		return
	}

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

// GetStockList 获取市场内指定范围内的所有证券代码
func (this *Client) GetStockList(exchange protocol.Exchange, starts ...uint16) (*protocol.StockListResp, error) {
	f := protocol.MStockList.Frame(exchange, starts...)
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.StockListResp), nil
}

// GetStockQuotes 获取盘口五档报价
func (this *Client) GetStockQuotes(m map[protocol.Exchange]string) (protocol.StockQuotesResp, error) {
	f, err := protocol.MStockQuote.Frame(m)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(protocol.StockQuotesResp), nil
}

// GetStockMinute 获取分时数据
func (this *Client) GetStockMinute(exchange protocol.Exchange, code string) (*protocol.StockMinuteResp, error) {
	f, err := protocol.MStockMinute.Frame(exchange, code)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.StockMinuteResp), nil
}
