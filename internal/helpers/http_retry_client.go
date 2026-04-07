package helpers

import (
	"net/http"
	"sync"

	"github.com/hashicorp/go-cleanhttp"
)

type HTTPRetryClient struct {
	client *http.Client

	RetryCount int

	clientInit sync.Once
}

func NewHTTPRetryClient() *HTTPRetryClient {
	return &HTTPRetryClient{client: &http.Client{}, RetryCount: 0, clientInit: sync.Once{}}
}

func (c *HTTPRetryClient) Do(req *http.Request) (*http.Response, error) {
	c.clientInit.Do(func() {
		if c.client == nil {
			c.client = cleanhttp.DefaultPooledClient()
		}
	})

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
