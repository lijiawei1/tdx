package protocol

type Control uint8

func (this Control) Uint8() uint8 {
	return uint8(this)
}

const (
	Control01 Control = 0x01 //好像都是01，暂时不知道啥含义
)

type Exchange uint8

func (this Exchange) Uint8() uint8 { return uint8(this) }

func (this Exchange) String() string {
	switch this {
	case ExchangeSH:
		return "sh"
	case ExchangeSZ:
		return "sz"
	case ExchangeBJ:
		return "bj"
	default:
		return "unknown"
	}
}

func (this Exchange) Name() string {
	switch this {
	case ExchangeSH:
		return "上海"
	case ExchangeSZ:
		return "深圳"
	case ExchangeBJ:
		return "北京"
	default:
		return "未知"
	}
}

const (
	ExchangeSH Exchange = iota //上海交易所
	ExchangeSZ                 //深圳交易所
	ExchangeBJ                 //北京交易所
)

type TypeKline uint8

func (this TypeKline) Uint16() uint16 { return uint16(this) }

const (
	TypeKline5Minute  TypeKline = 0  // 5分钟K 线
	TypeKline15Minute TypeKline = 1  // 15分钟K 线
	TypeKline30Minute TypeKline = 2  // 30分钟K 线
	TypeKlineHour     TypeKline = 3  // 1小时K 线
	TypeKlineDay2     TypeKline = 4  // 日K 线
	TypeKlineWeek     TypeKline = 5  // 周K 线
	TypeKlineMonth    TypeKline = 6  // 月K 线
	TypeKlineMinute   TypeKline = 7  // 1分钟
	TypeKlineMinute2  TypeKline = 8  // 1分钟K 线
	TypeKlineDay      TypeKline = 9  // 日K 线
	TypeKlineQuarter  TypeKline = 10 // 季K 线
	TypeKlineYear     TypeKline = 11 // 年K 线
)
