package main

import (
	"context"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/extend"
	"path/filepath"
)

func main() {

	m, err := tdx.NewManage(nil)
	logs.PanicErr(err)

	err = extend.NewPullKline(
		[]string{"sz000001"},
		[]string{extend.Year},
		filepath.Join(tdx.DefaultDatabaseDir, "kline"),
		1,
	).Run(context.Background(), m)
	logs.PanicErr(err)

}
