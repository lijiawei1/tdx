package extend

import (
	"context"
	_ "github.com/glebarez/go-sqlite"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
	"path/filepath"
	"sort"
	"xorm.io/core"
	"xorm.io/xorm"
)

const (
	Minute   = "minute"
	Minute5  = "5minute"
	Minute15 = "15minute"
	Minute30 = "30minute"
	Hour     = "hour"
	Day      = "day"
	Week     = "week"
	Month    = "month"
	Quarter  = "quarter"
	Year     = "year"
)

var (
	AllKlineType  = []string{Minute, Minute5, Minute15, Minute30, Hour, Day, Week, Month, Quarter, Year}
	KlineTableMap = map[string]*KlineTable{
		Minute:   NewKlineTable("MinuteKline", func(c *tdx.Client) KlineHandler { return c.GetKlineMinuteUntil }),
		Minute5:  NewKlineTable("Minute5Kline", func(c *tdx.Client) KlineHandler { return c.GetKline5MinuteUntil }),
		Minute15: NewKlineTable("Minute15Kline", func(c *tdx.Client) KlineHandler { return c.GetKline15MinuteUntil }),
		Minute30: NewKlineTable("Minute30Kline", func(c *tdx.Client) KlineHandler { return c.GetKline30MinuteUntil }),
		Hour:     NewKlineTable("HourKline", func(c *tdx.Client) KlineHandler { return c.GetKlineHourUntil }),
		Day:      NewKlineTable("DayKline", func(c *tdx.Client) KlineHandler { return c.GetKlineDayUntil }),
		Week:     NewKlineTable("WeekKline", func(c *tdx.Client) KlineHandler { return c.GetKlineWeekUntil }),
		Month:    NewKlineTable("MonthKline", func(c *tdx.Client) KlineHandler { return c.GetKlineMonthUntil }),
		Quarter:  NewKlineTable("QuarterKline", func(c *tdx.Client) KlineHandler { return c.GetKlineQuarterUntil }),
		Year:     NewKlineTable("YearKline", func(c *tdx.Client) KlineHandler { return c.GetKlineYearUntil }),
	}
)

func NewPullKline(codes, tables []string, dir string, limit int) *PullKline {
	_tables := []*KlineTable(nil)
	for _, v := range tables {
		_tables = append(_tables, KlineTableMap[v])
	}
	return &PullKline{
		tables: _tables,
		dir:    dir,
		Codes:  codes,
		limit:  limit,
	}
}

type PullKline struct {
	tables []*KlineTable
	dir    string   //数据目录
	Codes  []string //指定的代码
	limit  int      //并发数量
}

func (this *PullKline) Name() string {
	return "拉取k线数据"
}

func (this *PullKline) Run(ctx context.Context, m *tdx.Manage) error {
	limit := chans.NewWaitLimit(uint(this.limit))

	//1. 获取所有股票代码
	codes := this.Codes
	if len(codes) == 0 {
		codes = m.Codes.GetStocks()
	}

	for _, v := range codes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		limit.Add()
		go func(code string) {
			defer limit.Done()

			//连接数据库
			db, err := xorm.NewEngine("sqlite", filepath.Join(this.dir, code+".db"))
			if err != nil {
				logs.Err(err)
				return
			}
			db.SetMapper(core.SameMapper{})
			db.DB().SetMaxOpenConns(1)

			for _, table := range this.tables {
				if table == nil {
					continue
				}

				select {
				case <-ctx.Done():
					return
				default:
				}

				db.Sync2(table)

				//2. 获取最后一条数据
				last := new(Kline)
				if _, err = db.Table(table).Desc("Date").Get(last); err != nil {
					logs.Err(err)
					return
				}

				//3. 从服务器获取数据
				insert := Klines{}
				err = m.Do(func(c *tdx.Client) error {
					insert, err = this.pull(code, last.Date, table.Handler(c))
					return err
				})
				if err != nil {
					logs.Err(err)
					return
				}

				//4. 插入数据库
				err = tdx.NewSessionFunc(db, func(session *xorm.Session) error {
					for i, v := range insert {
						if i == 0 {
							if _, err := session.Table(table).Where("Date >= ?", v.Date).Delete(); err != nil {
								return err
							}
						}
						if _, err := session.Table(table).Insert(v); err != nil {
							return err
						}
					}
					return nil
				})
				logs.PrintErr(err)

			}

		}(v)
	}
	limit.Wait()
	return nil
}

func (this *PullKline) pull(code string, lastDate int64, f func(code string, f func(k *protocol.Kline) bool) (*protocol.KlineResp, error)) (Klines, error) {

	if lastDate == 0 {
		lastDate = protocol.ExchangeEstablish.Unix()
	}

	resp, err := f(code, func(k *protocol.Kline) bool {
		return k.Time.Unix() <= lastDate
	})
	if err != nil {
		return nil, err
	}

	ks := Klines{}
	for _, v := range resp.List {
		ks = append(ks, &Kline{
			Code:   code,
			Date:   v.Time.Unix(),
			Open:   v.Open,
			High:   v.High,
			Low:    v.Low,
			Close:  v.Close,
			Volume: v.Volume,
			Amount: v.Amount,
		})
	}

	return ks, nil
}

type Kline struct {
	Code   string         `json:"code" xorm:"-"`         //代码
	Date   int64          `json:"date"`                  //时间节点 2006-01-02 15:00
	Open   protocol.Price `json:"open"`                  //开盘价
	High   protocol.Price `json:"high"`                  //最高价
	Low    protocol.Price `json:"low"`                   //最低价
	Close  protocol.Price `json:"close"`                 //收盘价
	Volume int64          `json:"volume"`                //成交量
	Amount protocol.Price `json:"amount"`                //成交额
	InDate int64          `json:"inDate" xorm:"created"` //创建时间
}

type Klines []*Kline

func (this Klines) Less(i, j int) bool { return this[i].Code > this[j].Code }

func (this Klines) Swap(i, j int) { this[i], this[j] = this[j], this[i] }

func (this Klines) Len() int { return len(this) }

func (this Klines) Sort() { sort.Sort(this) }

// Kline 计算多个K线,成一个K线
func (this Klines) Kline() *Kline {
	if this == nil {
		return new(Kline)
	}
	k := new(Kline)
	for i, v := range this {
		switch i {
		case 0:
			k.Open = v.Open
			k.High = v.High
			k.Low = v.Low
			k.Close = v.Close
		case len(this) - 1:
			k.Close = v.Close
			k.Date = v.Date
		}
		if v.High > k.High {
			k.High = v.High
		}
		if v.Low < k.Low {
			k.Low = v.Low
		}
		k.Volume += v.Volume
		k.Amount += v.Amount
	}

	return k
}

// Merge 合并K线
func (this Klines) Merge(n int) Klines {
	if this == nil {
		return nil
	}
	ks := []*Kline(nil)
	for i := 0; i < len(this); i += n {
		if i+n > len(this) {
			ks = append(ks, this[i:].Kline())
		} else {
			ks = append(ks, this[i:i+n].Kline())
		}
	}
	return ks
}

type KlineHandler func(code string, f func(k *protocol.Kline) bool) (*protocol.KlineResp, error)

func NewKlineTable(tableName string, handler func(c *tdx.Client) KlineHandler) *KlineTable {
	return &KlineTable{
		tableName: tableName,
		Handler:   handler,
	}
}

type KlineTable struct {
	Kline     `xorm:"extends"`
	tableName string
	Handler   func(c *tdx.Client) KlineHandler `xorm:"-"`
}

func (this *KlineTable) TableName() string {
	return this.tableName
}
