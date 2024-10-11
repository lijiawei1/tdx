package tdx

import "net"

func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn: conn,
	}, nil
}

type Client struct {
	conn net.Conn
}

func (this *Client) Close() error {
	return this.conn.Close()
}

// GetSecurityList 获取市场内指定范围内的所有证券代码
func (this *Client) GetSecurityList() {

}
