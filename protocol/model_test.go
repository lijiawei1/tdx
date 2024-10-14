package protocol

import (
	"testing"
)

/*
0c00000000011a001a003e05050000000000000002000030303030303101363030303038
0c02000000011a001a003e05050000000000000002000030303030303101363030303038
*/
func TestNewSecurityQuotes(t *testing.T) {
	f, err := NewSecurityQuotes(map[Exchange]string{
		ExchangeSH: "000001",
		ExchangeSZ: "600008",
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(f.Bytes().HEX())
}
