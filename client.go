package tdx

import (
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/maps/wait/v2"
	"github.com/injoyai/conv"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx/protocol"
	"sync/atomic"
	"time"
)

// WithDebug 是否打印通讯数据
func WithDebug(b ...bool) client.Option {
	return func(c *client.Client) {
		c.Logger.Debug(b...)
	}
}

// WithRedial 断线重连
func WithRedial(b ...bool) client.Option {
	return func(c *client.Client) {
		c.SetRedial(b...)
	}
}

// Dial 与服务器建立连接
func Dial(addr string, op ...client.Option) (cli *Client, err error) {

	cli = &Client{
		Wait: wait.New(time.Second * 2),
		m:    maps.NewSafe(),
	}

	cli.Client, err = dial.TCP(addr, func(c *client.Client) {
		c.Logger.Debug(true)                           //开启日志打印
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

		f := protocol.MConnect.Frame()
		if _, err = c.Write(f.Bytes()); err != nil {
			c.Close()
		}
	})
	if err != nil {
		return nil, err
	}

	go cli.Client.Run()

	return cli, nil
}

type Client struct {
	*client.Client              //客户端实例
	Wait           *wait.Entity //异步回调,设置超时时间,超时则返回错误
	m              *maps.Safe   //有部分解析需要用到代码,返回数据获取不到,固请求的时候缓存下
	msgID          uint32       //消息id,使用SendFrame自动累加
}

// handlerDealMessage 处理服务器响应的数据
func (this *Client) handlerDealMessage(c *client.Client, msg ios.Acker) {

	defer func() {
		if e := recover(); e != nil {
			logs.Err(e)
		}
	}()

	f, err := protocol.Decode(msg.Payload())
	if err != nil {
		logs.Err(err)
		return
	}

	val, _ := this.m.GetAndDel(conv.String(f.MsgID))

	var resp any
	switch f.Type {

	case protocol.TypeConnect:

	case protocol.TypeHeart:

	case protocol.TypeStockCount:
		resp, err = protocol.MStockCount.Decode(f.Data)

	case protocol.TypeStockList:
		resp, err = protocol.MStockList.Decode(f.Data)

	case protocol.TypeStockQuote:
		resp = protocol.MStockQuote.Decode(f.Data)

	case protocol.TypeStockMinute:
		resp, err = protocol.MStockMinute.Decode(f.Data)

	case protocol.TypeStockMinuteTrade:
		resp, err = protocol.MStockMinuteTrade.Decode(f.Data, conv.String(val)) //todo

	case protocol.TypeStockHistoryMinuteTrade:
		resp, err = protocol.MStockHistoryMinuteTrade.Decode(f.Data, conv.String(val))

	case protocol.TypeStockKline:
		resp, err = protocol.MStockKline.Decode(f.Data, protocol.TypeKline(conv.Uint16(val)))

	default:
		err = fmt.Errorf("通讯类型未解析:0x%X", f.Type)

	}

	if err != nil {
		logs.Err(err)
		return
	}

	this.Wait.Done(conv.String(f.MsgID), resp)

}

// SendFrame 发送数据,并等待响应
func (this *Client) SendFrame(f *protocol.Frame) (any, error) {
	f.MsgID = atomic.AddUint32(&this.msgID, 1)
	if _, err := this.Client.Write(f.Bytes()); err != nil {
		return nil, err
	}
	return this.Wait.Wait(conv.String(this.msgID))
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
// 600开头的股票是上证A股，属于大盘股，其中6006开头的股票是最早上市的股票， 6016开头的股票为大盘蓝筹股；900开头的股票是上证B股；
// 000开头的股票是深证A股，001、002开头的股票也都属于深证A股， 其中002开头的股票是深证A股中小企业股票；200开头的股票是深证B股；
// 300开头的股票是创业板股票；400开头的股票是三板市场股票。
func (this *Client) GetStockList(exchange protocol.Exchange, start uint16) (*protocol.StockListResp, error) {
	f := protocol.MStockList.Frame(exchange, start)
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.StockListResp), nil
}

// GetStockAll 通过多次请求的方式获取全部证券代码
func (this *Client) GetStockAll(exchange protocol.Exchange) (*protocol.StockListResp, error) {
	resp := &protocol.StockListResp{}
	maxSize := uint16(1000)
	for start := uint16(0); ; start += maxSize {
		r, err := this.GetStockList(exchange, start)
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

// GetStockKline 获取k线数据
func (this *Client) GetStockKline(Type protocol.TypeKline, req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	f, err := protocol.MStockKline.Frame(Type, req)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.StockKlineResp), nil
}

// GetStockKlineMinute 获取一分钟k线数据
func (this *Client) GetStockKlineMinute(req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	return this.GetStockKline(protocol.TypeKlineMinute, req)
}

// GetStockKline5Minute 获取五分钟k线数据
func (this *Client) GetStockKline5Minute(req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	return this.GetStockKline(protocol.TypeKline5Minute, req)
}

// GetStockKline15Minute 获取十五分钟k线数据
func (this *Client) GetStockKline15Minute(req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	return this.GetStockKline(protocol.TypeKline15Minute, req)
}

// GetStockKline30Minute 获取三十分钟k线数据
func (this *Client) GetStockKline30Minute(req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	return this.GetStockKline(protocol.TypeKline30Minute, req)
}

// GetStockKlineHour 获取小时k线数据
func (this *Client) GetStockKlineHour(req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	return this.GetStockKline(protocol.TypeKlineHour, req)
}

// GetStockKlineDay 获取日k线数据
func (this *Client) GetStockKlineDay(req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	return this.GetStockKline(protocol.TypeKlineDay, req)
}

// GetStockKlineWeek 获取周k线数据
func (this *Client) GetStockKlineWeek(req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	return this.GetStockKline(protocol.TypeKlineWeek, req)
}

// GetStockKlineMonth 获取月k线数据
func (this *Client) GetStockKlineMonth(req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	return this.GetStockKline(protocol.TypeKlineMonth, req)
}

// GetStockKlineQuarter 获取季k线数据
func (this *Client) GetStockKlineQuarter(req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	return this.GetStockKline(protocol.TypeKlineQuarter, req)
}

// GetStockKlineYear 获取年k线数据
func (this *Client) GetStockKlineYear(req *protocol.StockKlineReq) (*protocol.StockKlineResp, error) {
	return this.GetStockKline(protocol.TypeKlineYear, req)
}
