package agent

import (
	"net/http"
)

type HTTPRetryClient struct {
	client *http.Client

	RetryCount int
}

func NewHTTPRetryClient() *HTTPRetryClient {
	client := &http.Client{}
	return &HTTPRetryClient{client: client}
}

func (c *HTTPRetryClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
