package panel

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

// Describe return a description of the client
func (c *Client) Describe() ClientInfo {
	return ClientInfo{APIHost: c.APIHost, NodeID: c.NodeID, Key: c.Key, NodeType: c.NodeType}
}

// Debug set the client debug for client
func (c *Client) Debug() {
	c.client.SetDebug(true)
}

func (c *Client) assembleURL(path string) string {
	return c.APIHost + path
}
func (c *Client) checkResponse(res *resty.Response, path string, err error) error {
	if err != nil {
		return fmt.Errorf("request %s failed: %s", c.assembleURL(path), err)
	}

	if res.StatusCode() > 400 {
		body := res.Body()
		return fmt.Errorf("request %s failed: %s, %s", c.assembleURL(path), string(body), err)
	}
	return nil
}
