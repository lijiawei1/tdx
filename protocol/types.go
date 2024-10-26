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
