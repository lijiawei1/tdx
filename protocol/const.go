package protocol

const (
	TypeConnect       = 0x000d //建立连接
	TypeHandshake     = 0xdb0f //握手
	TypeSecurityList  = 0x0450 //获取股票代码
	TypeSecurityQuote = 0x053e // 行情信息
)

const (
	LOGIN_ONE       = 0x000d //第一次登录
	LOGIN_TWO       = 0x0fdb //第二次登录
	HEART           = 0x0004 //心跳维持
	STOCK_COUNT     = 0x044e //股票数目
	STOCK_LIST      = 0x0450 //股票列表
	KMINUTE         = 0x0537 //当天分时K线
	KMINUTE_OLD     = 0x0fb4 //指定日期分时K线
	KLINE           = 0x052d //股票K线
	BIDD            = 0x056a //当日的竞价
	QUOTE           = 0x053e //实时五笔报价
	QUOTE_SORT      = 0x053e //沪深排序
	TRANSACTION     = 0x0fc5 //分笔成交明细
	TRANSACTION_OLD = 0x0fb5 //历史分笔成交明细
	FINANCE         = 0x0010 //财务数据
	COMPANY         = 0x02d0 //公司数据  F10
	EXDIVIDEND      = 0x000f //除权除息
	FILE_DIRECTORY  = 0x02cf //公司文件目录
	FILE_CONTENT    = 0x02d0 //公司文件内容
)

const (
	KMSG_CMD1                   = 0x000d // 建立链接
	KMSG_CMD2                   = 0x0fdb // 建立链接
	KMSG_PING                   = 0x0015 // 测试连接
	KMSG_HEARTBEAT              = 0xFFFF // 心跳(自定义)
	KMSG_SECURITYCOUNT          = 0x044e // 证券数量
	KMSG_BLOCKINFOMETA          = 0x02c5 // 板块文件信息
	KMSG_BLOCKINFO              = 0x06b9 // 板块文件
	KMSG_COMPANYCATEGORY        = 0x02cf // 公司信息文件信息
	KMSG_COMPANYCONTENT         = 0x02d0 // 公司信息描述
	KMSG_FINANCEINFO            = 0x0010 // 财务信息
	KMSG_HISTORYMINUTETIMEDATE  = 0x0fb4 // 历史分时信息
	KMSG_HISTORYTRANSACTIONDATA = 0x0fb5 // 历史分笔成交信息
	KMSG_INDEXBARS              = 0x052d // 指数K线
	KMSG_SECURITYBARS           = 0x052d // 股票K线
	KMSG_MINUTETIMEDATA         = 0x0537 // 分时数据
	KMSG_SECURITYLIST           = 0x0450 // 证券列表
	KMSG_SECURITYQUOTES         = 0x053e // 行情信息
	KMSG_TRANSACTIONDATA        = 0x0fc5 // 分笔成交信息
	KMSG_XDXRINFO               = 0x000f // 除权除息信息

)

const (
	KLINE_TYPE_5MIN      = 0  // 5分钟K 线
	KLINE_TYPE_15MIN     = 1  // 15分钟K 线
	KLINE_TYPE_30MIN     = 2  // 30分钟K 线
	KLINE_TYPE_1HOUR     = 3  // 1小时K 线
	KLINE_TYPE_DAILY     = 4  // 日K 线
	KLINE_TYPE_WEEKLY    = 5  // 周K 线
	KLINE_TYPE_MONTHLY   = 6  // 月K 线
	KLINE_TYPE_EXHQ_1MIN = 7  // 1分钟
	KLINE_TYPE_1MIN      = 8  // 1分钟K 线
	KLINE_TYPE_RI_K      = 9  // 日K 线
	KLINE_TYPE_3MONTH    = 10 // 季K 线
	KLINE_TYPE_YEARLY    = 11 // 年K 线
)
