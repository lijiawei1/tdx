package tdx

import (
	"context"
	"github.com/injoyai/ios"
	"net"
	"strings"
	"time"
)

func NewHostDial(hosts []string, timeout time.Duration) ios.DialFunc {
	if len(hosts) == 0 {
		hosts = Hosts
	}
	index := 0

	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		defer func() { index++ }()
		if index >= len(hosts) {
			index = 0
		}
		addr := hosts[index]
		if !strings.Contains(addr, ":") {
			addr += ":7709"
		}
		c, err := net.DialTimeout("tcp", addr, timeout)
		return c, hosts[index], err
	}
}
