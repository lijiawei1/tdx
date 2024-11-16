package tdx

import (
	"context"
	"github.com/injoyai/ios"
	"net"
	"strings"
)

func NewHostDial(hosts []string) ios.DialFunc {
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
		c, err := net.Dial("tcp", addr)
		return c, hosts[index], err
	}
}
