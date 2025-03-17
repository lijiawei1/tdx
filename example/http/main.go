package main

import (
	"github.com/injoyai/tdx"
)

func main() {
	// 启动 HTTP 服务器，监听 8080 端口
	go tdx.StartHTTPServer(8080)

	// 保持主函数运行
	select {}
}
