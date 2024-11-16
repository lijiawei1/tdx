package tdx

import (
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/maps/wait/v2"
	"github.com/injoyai/conv"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/module/tcp"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx/protocol"
	"runtime/debug"
	"strings"
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
	if !strings.Contains(addr, ":") {
		addr += ":7709"
	}
	return DialWith(tcp.NewDial(addr), op...)
}

// DialWith 与服务器建立连接
func DialWith(dial ios.DialFunc, op ...client.Option) (cli *Client, err error) {

	cli = &Client{
		Wait: wait.New(time.Second * 2),
		m:    maps.NewSafe(),
	}

	cli.Client, err = client.Dial(dial, func(c *client.Client) {
		c.Logger.Debug(false)                          //关闭日志打印
		c.Logger.WithHEX()                             //以HEX显示
		c.SetOption(op...)                             //自定义选项
		c.Event.OnReadFrom = protocol.ReadFrom         //分包
		c.Event.OnDealMessage = cli.handlerDealMessage //解析数据并处理
		//无数据超时时间是60秒,30秒发送一个心跳包
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
			debug.PrintStack()
		}
	}()

	f, err := protocol.Decode(msg.Payload())
	if err != nil {
		logs.Err(err)
		return
	}

	//从缓存中获取数据,响应数据中不同类型有不同的处理方式,但是响应无返回该类型,固根据消息id进行缓存
	val, _ := this.m.GetAndDel(conv.String(f.MsgID))

	var resp any
	switch f.Type {

	case protocol.TypeConnect:

	case protocol.TypeHeart:

	case protocol.TypeCount:
		resp, err = protocol.MCount.Decode(f.Data)

	case protocol.TypeCode:
		resp, err = protocol.MCode.Decode(f.Data)

	case protocol.TypeQuote:
		resp = protocol.MQuote.Decode(f.Data)

	case protocol.TypeMinute:
		resp, err = protocol.MMinute.Decode(f.Data)

	case protocol.TypeMinuteTrade:
		resp, err = protocol.MMinuteTrade.Decode(f.Data, conv.String(val)) //todo

	case protocol.TypeHistoryMinuteTrade:
		resp, err = protocol.MHistoryMinuteTrade.Decode(f.Data, conv.String(val))

	case protocol.TypeKline:
		resp, err = protocol.MKline.Decode(f.Data, conv.Uint8(val))

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
func (this *Client) SendFrame(f *protocol.Frame, cache ...any) (any, error) {
	f.MsgID = atomic.AddUint32(&this.msgID, 1)
	if len(cache) > 0 {
		this.m.Set(conv.String(f.MsgID), cache[0])
	}
	if _, err := this.Client.Write(f.Bytes()); err != nil {
		return nil, err
	}
	return this.Wait.Wait(conv.String(this.msgID))
}

// GetCount 获取市场内的股票数量
func (this *Client) GetCount(exchange protocol.Exchange) (*protocol.CountResp, error) {
	f := protocol.MCount.Frame(exchange)
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.CountResp), nil
}

// GetCode 获取市场内指定范围内的所有证券代码,一次固定返回1000只,上证股票有效范围370-1480
// 上证前370只是395/399开头的(中证500/总交易等辅助类),在后面的话是一些100开头的国债
// 600开头的股票是上证A股，属于大盘股，其中6006开头的股票是最早上市的股票， 6016开头的股票为大盘蓝筹股；900开头的股票是上证B股；
// 000开头的股票是深证A股，001、002开头的股票也都属于深证A股， 其中002开头的股票是深证A股中小企业股票；200开头的股票是深证B股；
// 300开头的股票是创业板股票；400开头的股票是三板市场股票。
func (this *Client) GetCode(exchange protocol.Exchange, start uint16) (*protocol.CodeResp, error) {
	f := protocol.MCode.Frame(exchange, start)
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.CodeResp), nil
}

// GetCodeAll 通过多次请求的方式获取全部证券代码
func (this *Client) GetCodeAll(exchange protocol.Exchange) (*protocol.CodeResp, error) {
	resp := &protocol.CodeResp{}
	size := uint16(1000)
	for start := uint16(0); ; start += size {
		r, err := this.GetCode(exchange, start)
		if err != nil {
			return nil, err
		}
		resp.Count += r.Count
		resp.List = append(resp.List, r.List...)
		if r.Count < size {
			break
		}
	}
	return resp, nil
}

// GetQuote 获取盘口五档报价
func (this *Client) GetQuote(codes ...string) (protocol.QuotesResp, error) {
	f, err := protocol.MQuote.Frame(codes...)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(protocol.QuotesResp), nil
}

// GetMinute 获取分时数据,todo 解析好像不对
func (this *Client) GetMinute(code string) (*protocol.MinuteResp, error) {
	f, err := protocol.MMinute.Frame(code)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.MinuteResp), nil
}

// GetMinuteTrade 获取分时交易详情,服务器最多返回1800条,count-start<=1800
func (this *Client) GetMinuteTrade(code string, start, count uint16) (*protocol.MinuteTradeResp, error) {
	f, err := protocol.MMinuteTrade.Frame(code, start, count)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f, code)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.MinuteTradeResp), nil
}

// GetMinuteTradeAll 获取分时全部交易详情,todo 只做参考 因为交易实时在进行,然后又是分页读取的,所以会出现读取间隔内产生的交易会丢失
func (this *Client) GetMinuteTradeAll(code string) (*protocol.MinuteTradeResp, error) {
	resp := &protocol.MinuteTradeResp{}
	size := uint16(1800)
	for start := uint16(0); ; start += size {
		r, err := this.GetMinuteTrade(code, start, size)
		if err != nil {
			return nil, err
		}
		resp.Count += r.Count
		resp.List = append(r.List, resp.List...)

		if r.Count < size {
			break
		}
	}
	return resp, nil
}

// GetHistoryMinuteTrade 获取历史分时交易
// 只能获取昨天及之前的数据,服务器最多返回2000条,count-start<=2000,如果日期输入错误,则返回0
// 历史数据sz000001在20241116只能查到21111112,13年差几天,3141天,或者其他规则
func (this *Client) GetHistoryMinuteTrade(date, code string, start, count uint16) (*protocol.HistoryMinuteTradeResp, error) {
	f, err := protocol.MHistoryMinuteTrade.Frame(date, code, start, count)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f, code)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.HistoryMinuteTradeResp), nil
}

// GetHistoryMinuteTradeAll 获取历史分时全部交易,通过多次请求来拼接,只能获取昨天及之前的数据
// 历史数据sz000001在20241116只能查到21111112,13年差几天,3141天,或者其他规则
func (this *Client) GetHistoryMinuteTradeAll(date, code string) (*protocol.HistoryMinuteTradeResp, error) {
	resp := &protocol.HistoryMinuteTradeResp{}
	size := uint16(2000)
	for start := uint16(0); ; start += size {
		r, err := this.GetHistoryMinuteTrade(date, code, start, size)
		if err != nil {
			return nil, err
		}
		resp.Count += r.Count
		resp.List = append(r.List, resp.List...)
		if r.Count < size {
			break
		}
	}
	return resp, nil
}

// GetKline 获取k线数据,推荐收盘之后获取,否则会获取到当天的数据
func (this *Client) GetKline(Type uint8, code string, start, count uint16) (*protocol.KlineResp, error) {
	f, err := protocol.MKline.Frame(Type, code, start, count)
	if err != nil {
		return nil, err
	}
	result, err := this.SendFrame(f, Type)
	if err != nil {
		return nil, err
	}
	return result.(*protocol.KlineResp), nil
}

// GetKlineAll 获取全部k线数据
func (this *Client) GetKlineAll(Type uint8, code string) (*protocol.KlineResp, error) {
	resp := &protocol.KlineResp{}
	size := uint16(800)
	var last *protocol.Kline
	for start := uint16(0); ; start += size {
		r, err := this.GetKline(Type, code, start, size)
		if err != nil {
			return nil, err
		}
		if last != nil && len(r.List) > 0 {
			last.Last = r.List[len(r.List)-1].Close
		}
		if len(r.List) > 0 {
			last = r.List[0]
		}
		resp.Count += r.Count
		resp.List = append(r.List, resp.List...)
		if r.Count < size {
			break
		}
	}
	return resp, nil
}

// GetKlineMinute 获取一分钟k线数据,每次最多800条,最多只能获取24000条数据
func (this *Client) GetKlineMinute(code string, start, count uint16) (*protocol.KlineResp, error) {
	return this.GetKline(protocol.TypeKlineMinute, code, start, count)
}

// GetKlineMinuteAll 获取一分钟k线全部数据,最多只能获取24000条数据
func (this *Client) GetKlineMinuteAll(code string) (*protocol.KlineResp, error) {
	return this.GetKlineAll(protocol.TypeKlineMinute, code)
}

// GetKline5Minute 获取五分钟k线数据
func (this *Client) GetKline5Minute(code string, start, count uint16) (*protocol.KlineResp, error) {
	return this.GetKline(protocol.TypeKline5Minute, code, start, count)
}

// GetKline5MinuteAll 获取5分钟k线全部数据
func (this *Client) GetKline5MinuteAll(code string) (*protocol.KlineResp, error) {
	return this.GetKlineAll(protocol.TypeKline5Minute, code)
}

// GetKline15Minute 获取十五分钟k线数据
func (this *Client) GetKline15Minute(code string, start, count uint16) (*protocol.KlineResp, error) {
	return this.GetKline(protocol.TypeKline15Minute, code, start, count)
}

// GetKline15MinuteAll 获取十五分钟k线全部数据
func (this *Client) GetKline15MinuteAll(code string) (*protocol.KlineResp, error) {
	return this.GetKlineAll(protocol.TypeKline15Minute, code)
}

// GetKline30Minute 获取三十分钟k线数据
func (this *Client) GetKline30Minute(code string, start, count uint16) (*protocol.KlineResp, error) {
	return this.GetKline(protocol.TypeKline30Minute, code, start, count)
}

// GetKline30MinuteAll 获取三十分钟k线全部数据
func (this *Client) GetKline30MinuteAll(code string) (*protocol.KlineResp, error) {
	return this.GetKlineAll(protocol.TypeKline30Minute, code)
}

// GetKlineHour 获取小时k线数据
func (this *Client) GetKlineHour(code string, start, count uint16) (*protocol.KlineResp, error) {
	return this.GetKline(protocol.TypeKlineHour, code, start, count)
}

// GetKlineHourAll 获取小时k线全部数据
func (this *Client) GetKlineHourAll(code string) (*protocol.KlineResp, error) {
	return this.GetKlineAll(protocol.TypeKlineHour, code)
}

// GetKlineDay 获取日k线数据
func (this *Client) GetKlineDay(code string, start, count uint16) (*protocol.KlineResp, error) {
	return this.GetKline(protocol.TypeKlineDay, code, start, count)
}

// GetKlineDayAll 获取日k线全部数据
func (this *Client) GetKlineDayAll(code string) (*protocol.KlineResp, error) {
	return this.GetKlineAll(protocol.TypeKlineDay, code)
}

// GetKlineWeek 获取周k线数据
func (this *Client) GetKlineWeek(code string, start, count uint16) (*protocol.KlineResp, error) {
	return this.GetKline(protocol.TypeKlineWeek, code, start, count)
}

// GetKlineWeekAll 获取周k线全部数据
func (this *Client) GetKlineWeekAll(code string) (*protocol.KlineResp, error) {
	return this.GetKlineAll(protocol.TypeKlineWeek, code)
}

// GetKlineMonth 获取月k线数据
func (this *Client) GetKlineMonth(code string, start, count uint16) (*protocol.KlineResp, error) {
	return this.GetKline(protocol.TypeKlineMonth, code, start, count)
}

// GetKlineMonthAll 获取月k线全部数据
func (this *Client) GetKlineMonthAll(code string) (*protocol.KlineResp, error) {
	return this.GetKlineAll(protocol.TypeKlineMonth, code)
}

// GetKlineQuarter 获取季k线数据
func (this *Client) GetKlineQuarter(code string, start, count uint16) (*protocol.KlineResp, error) {
	return this.GetKline(protocol.TypeKlineQuarter, code, start, count)
}

// GetKlineQuarterAll 获取季k线全部数据
func (this *Client) GetKlineQuarterAll(code string) (*protocol.KlineResp, error) {
	return this.GetKlineAll(protocol.TypeKlineQuarter, code)
}

// GetKlineYear 获取年k线数据
func (this *Client) GetKlineYear(code string, start, count uint16) (*protocol.KlineResp, error) {
	return this.GetKline(protocol.TypeKlineYear, code, start, count)
}

// GetKlineYearAll 获取年k线数据
func (this *Client) GetKlineYearAll(code string) (*protocol.KlineResp, error) {
	return this.GetKlineAll(protocol.TypeKlineYear, code)
}
