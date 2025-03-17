package tdx

import (
	"encoding/json"
	"fmt"
	"github.com/injoyai/tdx/protocol"
	"log"
	"net/http"
	"strconv"
	"time"
)

// StartHTTPServer 启动 HTTP 服务器
func StartHTTPServer(port int) {
	http.HandleFunc("/get_quote", getQuoteHandler)
	http.HandleFunc("/get_count", getCountHandler)
	http.HandleFunc("/get_code", getCodeHandler)
	http.HandleFunc("/get_code_all", getCodeAllHandler)
	http.HandleFunc("/get_minute", getMinuteHandler)
	http.HandleFunc("/get_history_minute", getHistoryMinuteHandler)
	http.HandleFunc("/get_minute_trade", getMinuteTradeHandler)
	http.HandleFunc("/get_minute_trade_all", getMinuteTradeAllHandler)
	http.HandleFunc("/get_history_minute_trade", getHistoryMinuteTradeHandler)
	http.HandleFunc("/get_history_minute_trade_all", getHistoryMinuteTradeAllHandler)
	http.HandleFunc("/get_index", getIndexHandler)
	http.HandleFunc("/get_kline_day", getKlineDayHandler)
	http.HandleFunc("/get_kline_day_all", getKlineDayAllHandler)
	http.HandleFunc("/get_kline_until", getKlineUntilHandler)
	http.HandleFunc("/get_kline_minute", getKlineMinuteHandler)
	http.HandleFunc("/get_kline_year", getKlineYearHandler)
	// 可以根据需要添加更多的接口处理函数

	address := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP server on port %d", port)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// getQuoteHandler 处理获取行情报价的请求
func getQuoteHandler(w http.ResponseWriter, r *http.Request) {
	codes := r.URL.Query()["code"]
	if len(codes) == 0 {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取行情报价
	resp, err := client.GetQuote(codes...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get quote: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getCountHandler 处理获取市场内股票数量的请求
func getCountHandler(w http.ResponseWriter, r *http.Request) {
	exchangeStr := r.URL.Query().Get("exchange")
	if exchangeStr == "" {
		http.Error(w, "Missing 'exchange' parameter", http.StatusBadRequest)
		return
	}

	// 将字符串转换为 Exchange 类型
	var exchange protocol.Exchange
	switch exchangeStr {
	case "sh":
		exchange = protocol.ExchangeSH
	case "sz":
		exchange = protocol.ExchangeSZ
	default:
		http.Error(w, "Invalid 'exchange' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取市场内股票数量
	resp, err := client.GetCount(exchange)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get count: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getCodeHandler 处理获取市场内指定范围内的所有证券代码的请求
func getCodeHandler(w http.ResponseWriter, r *http.Request) {
	exchangeStr := r.URL.Query().Get("exchange")
	if exchangeStr == "" {
		http.Error(w, "Missing 'exchange' parameter", http.StatusBadRequest)
		return
	}
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "Missing 'start' parameter", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(startStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'start' parameter", http.StatusBadRequest)
		return
	}

	// 将字符串转换为 Exchange 类型
	var exchange protocol.Exchange
	switch exchangeStr {
	case "sh":
		exchange = protocol.ExchangeSH
	case "sz":
		exchange = protocol.ExchangeSZ
	default:
		http.Error(w, "Invalid 'exchange' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取市场内指定范围内的所有证券代码
	resp, err := client.GetCode(exchange, uint16(start))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get code: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getCodeAllHandler 处理获取全部证券代码的请求
func getCodeAllHandler(w http.ResponseWriter, r *http.Request) {
	exchangeStr := r.URL.Query().Get("exchange")
	if exchangeStr == "" {
		http.Error(w, "Missing 'exchange' parameter", http.StatusBadRequest)
		return
	}

	// 将字符串转换为 Exchange 类型
	var exchange protocol.Exchange
	switch exchangeStr {
	case "sh":
		exchange = protocol.ExchangeSH
	case "sz":
		exchange = protocol.ExchangeSZ
	default:
		http.Error(w, "Invalid 'exchange' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取全部证券代码
	resp, err := client.GetCodeAll(exchange)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get code all: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getMinuteHandler 处理获取分时数据的请求
func getMinuteHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取分时数据
	resp, err := client.GetMinute(code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get minute: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getHistoryMinuteHandler 处理获取历史分时数据的请求
func getHistoryMinuteHandler(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "Missing 'date' parameter", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取历史分时数据
	resp, err := client.GetHistoryMinute(date, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get history minute: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getMinuteTradeHandler 处理获取分时交易详情的请求
func getMinuteTradeHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "Missing 'start' parameter", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(startStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'start' parameter", http.StatusBadRequest)
		return
	}
	countStr := r.URL.Query().Get("count")
	if countStr == "" {
		http.Error(w, "Missing 'count' parameter", http.StatusBadRequest)
		return
	}
	count, err := strconv.ParseUint(countStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'count' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取分时交易详情
	resp, err := client.GetMinuteTrade(code, uint16(start), uint16(count))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get minute trade: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getMinuteTradeAllHandler 处理获取分时全部交易详情的请求
func getMinuteTradeAllHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取分时全部交易详情
	resp, err := client.GetMinuteTradeAll(code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get minute trade all: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getHistoryMinuteTradeHandler 处理获取历史分时交易的请求
func getHistoryMinuteTradeHandler(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "Missing 'date' parameter", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "Missing 'start' parameter", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(startStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'start' parameter", http.StatusBadRequest)
		return
	}
	countStr := r.URL.Query().Get("count")
	if countStr == "" {
		http.Error(w, "Missing 'count' parameter", http.StatusBadRequest)
		return
	}
	count, err := strconv.ParseUint(countStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'count' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取历史分时交易
	resp, err := client.GetHistoryMinuteTrade(date, code, uint16(start), uint16(count))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get history minute trade: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getHistoryMinuteTradeAllHandler 处理获取历史分时全部交易的请求
func getHistoryMinuteTradeAllHandler(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "Missing 'date' parameter", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取历史分时全部交易
	resp, err := client.GetHistoryMinuteTradeAll(date, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get history minute trade all: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getIndexHandler 处理获取指数数据的请求
func getIndexHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取指数数据
	resp, err := client.GetKlineDay(code, 0, 10)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get index: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getKlineDayHandler 处理获取日K线数据的请求
func getKlineDayHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "Missing 'start' parameter", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(startStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'start' parameter", http.StatusBadRequest)
		return
	}
	countStr := r.URL.Query().Get("count")
	if countStr == "" {
		http.Error(w, "Missing 'count' parameter", http.StatusBadRequest)
		return
	}
	count, err := strconv.ParseUint(countStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'count' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取日K线数据
	resp, err := client.GetKlineDay(code, uint16(start), uint16(count))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get day kline: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getKlineDayAllHandler 处理获取全部日K线数据的请求
func getKlineDayAllHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取全部日K线数据
	resp, err := client.GetKlineDayAll(code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get all day kline: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getKlineUntilHandler 处理获取直到指定时间的K线数据的请求
func getKlineUntilHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}
	endDateStr := r.URL.Query().Get("end_date")
	if endDateStr == "" {
		http.Error(w, "Missing 'end_date' parameter", http.StatusBadRequest)
		return
	}
	countStr := r.URL.Query().Get("count")
	if countStr == "" {
		http.Error(w, "Missing 'count' parameter", http.StatusBadRequest)
		return
	}
	//count, err := strconv.ParseUint(countStr, 10, 16)
	//if err != nil {
	//	http.Error(w, "Invalid 'count' parameter", http.StatusBadRequest)
	//	return
	//}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	old := time.Now().Add(-time.Hour * 24 * 17)
	// 获取直到指定时间的K线数据
	resp, err := client.GetKlineUntil(protocol.TypeKlineDay, code, func(k *protocol.Kline) bool {
		return k.Time.Sub(old) < 0
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get kline until: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getKlineMinuteHandler 处理获取分钟K线数据的请求
func getKlineMinuteHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "Missing 'start' parameter", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(startStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'start' parameter", http.StatusBadRequest)
		return
	}
	countStr := r.URL.Query().Get("count")
	if countStr == "" {
		http.Error(w, "Missing 'count' parameter", http.StatusBadRequest)
		return
	}
	count, err := strconv.ParseUint(countStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'count' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取分钟K线数据
	resp, err := client.GetKlineMinute(code, uint16(start), uint16(count))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get minute kline: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}

// getKlineYearHandler 处理获取年K线数据的请求
func getKlineYearHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing 'code' parameter", http.StatusBadRequest)
		return
	}
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "Missing 'start' parameter", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(startStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'start' parameter", http.StatusBadRequest)
		return
	}
	countStr := r.URL.Query().Get("count")
	if countStr == "" {
		http.Error(w, "Missing 'count' parameter", http.StatusBadRequest)
		return
	}
	count, err := strconv.ParseUint(countStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid 'count' parameter", http.StatusBadRequest)
		return
	}

	// 连接服务器
	client, err := Dial("124.71.187.122:7709")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 获取年K线数据
	resp, err := client.GetKlineYear(code, uint16(start), uint16(count))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get year kline: %v", err), http.StatusInternalServerError)
		return
	}

	// 将响应数据转换为 JSON 格式
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	// 发送响应数据
	w.Write(jsonData)
}
