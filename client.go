package tdx

import (
	"github.com/injoyai/base/maps"
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
		m: maps.NewSafe(),
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
	m     *maps.Safe
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

	val, _ := this.m.GetAndDel(conv.String(f.MsgID))
	code := conv.String(val)

	var resp any
	switch f.Type {

	case protocol.TypeConnect:

	case protocol.TypeStockCount:
		resp, err = protocol.MStockCount.Decode(f.Data)

	case protocol.TypeStockList:
		resp, err = protocol.MStockList.Decode(f.Data)

	case protocol.TypeStockQuote:
		resp = protocol.MStockQuote.Decode(f.Data)

	case protocol.TypeStockMinute:
		resp, err = protocol.MStockMinute.Decode(f.Data)

	case protocol.TypeStockMinuteTrade:
		resp, err = protocol.MStockMinuteTrade.Decode(f.Data, code) //todo

	case protocol.TypeStockHistoryMinuteTrade:
		resp, err = protocol.MStockHistoryMinuteTrade.Decode(f.Data, code)

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

// GetStockCount 获取市场内的股票数量
func (this *Client) GetStockCount(exchange protocol.Exchange) (*protocol.StockCountResp, error) {
	f := protocol.MStockCount.Frame(exchange)
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.StockCountResp), nil
}

// GetStockList 获取市场内指定范围内的所有证券代码,一次固定返回1000只,上证股票有效范围370-1480
// 上证前370只是395/399开头的(中证500/总交易等辅助类),在后面的话是一些100开头的国债
func (this *Client) GetStockList(exchange protocol.Exchange, start uint16) (*protocol.StockListResp, error) {
	f := protocol.MStockList.Frame(exchange, start)
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

// GetStockMinute 获取分时数据,todo 解析好像不对
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

// GetStockMinuteTrade 获取分时交易详情,服务器最多返回1800条,count-start<=1800
func (this *Client) GetStockMinuteTrade(exchange protocol.Exchange, code string, start, count uint16) (*protocol.StockMinuteTradeResp, error) {
	f, err := protocol.MStockMinuteTrade.Frame(exchange, code, start, count)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.StockMinuteTradeResp), nil
}

// GetStockMinuteTradeAll 获取分时全部交易详情,todo 只做参考 因为交易实时在进行,然后又是分页读取的,所以会出现读取间隔内产生的交易会丢失
func (this *Client) GetStockMinuteTradeAll(exchange protocol.Exchange, code string) (*protocol.StockMinuteTradeResp, error) {
	resp := &protocol.StockMinuteTradeResp{}
	maxSize := uint16(1800)
	for i := uint16(0); ; i += maxSize {
		r, err := this.GetStockMinuteTrade(exchange, code, i, i+maxSize)
		if err != nil {
			return nil, err
		}
		resp.Count += r.Count
		resp.List = append(resp.List, r.List...)

		if r.Count < maxSize {
			break
		}
	}
	return resp, nil
}

// GetStockHistoryMinuteTrade 获取历史分时交易,,只能获取昨天及之前的数据,服务器最多返回2000条,count-start<=2000
func (this *Client) GetStockHistoryMinuteTrade(t time.Time, exchange protocol.Exchange, code string, start, count uint16) (*protocol.StockHistoryMinuteTradeResp, error) {
	f, err := protocol.MStockHistoryMinuteTrade.Frame(t, exchange, code, start, count)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.StockHistoryMinuteTradeResp), nil
}

// GetStockHistoryMinuteTradeAll 获取历史分时全部交易,通过多次请求来拼接,只能获取昨天及之前的数据
func (this *Client) GetStockHistoryMinuteTradeAll(exchange protocol.Exchange, code string) (*protocol.StockMinuteTradeResp, error) {
	resp := &protocol.StockMinuteTradeResp{}
	maxSize := uint16(2000)
	for i := uint16(0); ; i += maxSize {
		r, err := this.GetStockMinuteTrade(exchange, code, i, i+maxSize)
		if err != nil {
			return nil, err
		}
		resp.Count += r.Count
		resp.List = append(resp.List, r.List...)
		if r.Count < maxSize {
			break
		}
	}
	return resp, nil
}
