package tdx

import (
	"context"
	"github.com/injoyai/ios"
	"math/rand"
	"net"
	"strings"
	"time"
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
		return c, addr, err
	}
}

func NewRandomDial(hosts []string) ios.DialFunc {
	if len(hosts) == 0 {
		hosts = Hosts
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		addr := hosts[r.Intn(len(hosts))]
		if !strings.Contains(addr, ":") {
			addr += ":7709"
		}
		c, err := net.Dial("tcp", addr)
		return c, addr, err
	}
}
