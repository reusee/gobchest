package gobchest

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

func (c *Client) Close() {
	c.Client.Close()
}

func (c *Client) Set(key string, value interface{}) error {
	var response Response
	err := c.Call("Chest.Set", Request{
		Type:  Set,
		Key:   key,
		Value: value,
	}, &response)
	return err
}

func (c *Client) Get(key string) (interface{}, error) {
	var response Response
	err := c.Call("Chest.Get", Request{
		Type: Get,
		Key:  key,
	}, &response)
	if err != nil {
		return nil, err
	}
	return response.Value, nil
}

func (c *Client) ListAppend(key string, values ...interface{}) error {
	var response Response
	err := c.Call("Chest.ListAppend", Request{
		Type:  ListAppend,
		Key:   key,
		Value: values,
	}, &response)
	return err
}
