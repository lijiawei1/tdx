package tdx

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx/protocol"
	"github.com/robfig/cron/v3"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
	"xorm.io/core"
	"xorm.io/xorm"
)

// 增加单例,部分数据需要通过Codes里面的信息计算
var (
	DefaultCodes *Codes
	codesOnce    sync.Once
)

func NewCodes(c *Client, filename string) (*Codes, error) {

	//如果文件夹不存在就创建
	dir, _ := filepath.Split(filename)
	_ = os.MkdirAll(dir, 0777)

	//连接数据库
	db, err := xorm.NewEngine("sqlite", filename)
	if err != nil {
		return nil, err
	}
	db.SetMapper(core.SameMapper{})
	db.DB().SetMaxOpenConns(1)
	if err := db.Sync2(new(CodeModel)); err != nil {
		return nil, err
	}
	if err := db.Sync2(new(UpdateModel)); err != nil {
		return nil, err
	}

	update := new(UpdateModel)
	{ //查询或者插入一条数据
		has, err := db.Get(update)
		if err != nil {
			return nil, err
		} else if !has {
			if _, err := db.Insert(update); err != nil {
				return nil, err
			}
		}
	}

	cc := &Codes{
		Client: c,
		db:     db,
		Codes:  nil,
	}

	{ //设置定时器,每天早上9点更新数据
		task := cron.New(cron.WithSeconds())
		task.AddFunc("0 0 9 * * *", func() {
			for i := 0; i < 3; i++ {
				if err := cc.Update(); err == nil {
					return
				}
				logs.Err(err)
				<-time.After(time.Minute * 5)
			}
		})
		task.Start()
	}

	{ //判断是否更新过,更新过则不更新
		now := time.Now()
		node := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.Local)
		updateTime := time.Unix(update.Time, 0)
		if now.Sub(node) > 0 {
			//当前时间在9点之后,且更新时间在9点之前,需要更新
			if updateTime.Sub(node) < 0 {
				return cc, cc.Update()
			}
		} else {
			//当前时间在9点之前,且更新时间在上个节点之前
			if updateTime.Sub(node.Add(time.Hour*24)) < 0 {
				return cc, cc.Update()
			}
		}
	}

	//从缓存中加载
	return cc, cc.Update(true)
}

type Codes struct {
	*Client                       //客户端
	db      *xorm.Engine          //数据库实例
	Codes   map[string]*CodeModel //股票缓存
}

// GetName 获取股票名称
func (this *Codes) GetName(code string) string {
	if v, ok := this.Codes[code]; ok {
		return v.Name
	}
	return "未知"
}

// GetStocks 获取股票代码,sh6xxx sz0xx sz30xx
func (this *Codes) GetStocks() []string {
	ls := []string(nil)
	for k, _ := range this.Codes {
		if protocol.IsStock(k) {
			ls = append(ls, k)
		}
	}
	return ls
}

func (this *Codes) Get(code string) *CodeModel {
	if v, ok := this.Codes[code]; ok {
		return v
	}
	return nil
}

func (this *Codes) Update(byCache ...bool) error {
	codes, err := this.Code(len(byCache) > 0 && byCache[0])
	if err != nil {
		return err
	}
	codeMap := make(map[string]*CodeModel)
	for _, code := range codes {
		codeMap[code.Exchange+code.Code] = code
	}
	this.Codes = codeMap
	//更新时间
	_, err = this.db.Update(&UpdateModel{Time: time.Now().Unix()})
	return err
}

// Code 更新股票并返回结果
func (this *Codes) Code(byDatabase bool) ([]*CodeModel, error) {

	//2. 查询数据库所有股票
	list := []*CodeModel(nil)
	if err := this.db.Find(&list); err != nil {
		return nil, err
	}

	//如果是从缓存读取,则返回结果
	if byDatabase {
		return list, nil
	}

	mCode := make(map[string]*CodeModel, len(list))
	for _, v := range list {
		mCode[v.Code] = v
	}

	//3. 从服务器获取所有股票代码
	insert := []*CodeModel(nil)
	update := []*CodeModel(nil)
	for _, exchange := range []protocol.Exchange{protocol.ExchangeSH, protocol.ExchangeSZ} {
		resp, err := this.Client.GetCodeAll(exchange)
		if err != nil {
			return nil, err
		}
		for _, v := range resp.List {
			if _, ok := mCode[v.Code]; ok {
				if mCode[v.Code].Name != v.Name {
					mCode[v.Code].Name = v.Name
					update = append(update, &CodeModel{
						Name:      v.Name,
						Code:      v.Code,
						Exchange:  exchange.String(),
						Multiple:  v.Multiple,
						Decimal:   v.Decimal,
						LastPrice: v.LastPrice,
					})
				}
			} else {
				code := &CodeModel{
					Name:      v.Name,
					Code:      v.Code,
					Exchange:  exchange.String(),
					Multiple:  v.Multiple,
					Decimal:   v.Decimal,
					LastPrice: v.LastPrice,
				}
				insert = append(insert, code)
				list = append(list, code)
			}
		}
	}

	//4. 插入或者更新数据库
	err := NewSessionFunc(this.db, func(session *xorm.Session) error {
		for _, v := range insert {
			if _, err := session.Insert(v); err != nil {
				return err
			}
		}
		for _, v := range update {
			if _, err := session.Where("Exchange=? and Code=? ", v.Exchange, v.Code).Cols("Name").Update(v); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return list, nil

}

type UpdateModel struct {
	Time int64 //更新时间
}

func (*UpdateModel) TableName() string {
	return "update"
}

type CodeModel struct {
	ID        int64   `json:"id"`                      //主键
	Name      string  `json:"name"`                    //名称,有时候名称会变,例STxxx
	Code      string  `json:"code" xorm:"index"`       //代码
	Exchange  string  `json:"exchange" xorm:"index"`   //交易所
	Multiple  uint16  `json:"multiple"`                //倍数
	Decimal   int8    `json:"decimal"`                 //小数位
	LastPrice float64 `json:"lastPrice"`               //昨收价格
	EditDate  int64   `json:"editDate" xorm:"updated"` //修改时间
	InDate    int64   `json:"inDate" xorm:"created"`   //创建时间
}

func (*CodeModel) TableName() string {
	return "codes"
}

func (this *CodeModel) Price(p protocol.Price) protocol.Price {
	return p * protocol.Price(math.Pow10(int(3-this.Decimal)))
}

func NewSessionFunc(db *xorm.Engine, fn func(session *xorm.Session) error) error {
	session := db.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		session.Rollback()
		return err
	}
	if err := fn(session); err != nil {
		session.Rollback()
		return err
	}
	if err := session.Commit(); err != nil {
		session.Rollback()
		return err
	}
	return nil
}
