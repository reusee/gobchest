package store

import "net/rpc"

type Client struct {
	*rpc.Client
}

func NewClient(addr string) (*Client, error) {
	rpcClient, err := rpc.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Client: rpcClient,
	}
	return client, nil
}

func (c *Client) Set(key string, value interface{}) error {
	var response Response
	err := c.Call("Store.Set", Request{
		Type:  Set,
		Key:   key,
		Value: value,
	}, &response)
	return err
}

func (c *Client) Close() {
	c.Client.Close()
}
