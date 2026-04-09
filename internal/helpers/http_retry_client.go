package helpers

import (
	"io"
	"net/http"
	"time"
)

type HTTPDoRequester interface {
	Do(req *http.Request) (*http.Response, error)
}

type HTTPRetryClient struct {
	client HTTPDoRequester

	MaxAttempts  int
	AttemptDelay time.Duration

	shouldRetryOnStatus func(statusCode int) bool
}

func NewHTTPRetryClient(shouldRetryOnStatus func(int) bool) *HTTPRetryClient {
	if shouldRetryOnStatus == nil {
		shouldRetryOnStatus = defaultShouldRetryStatus
	}
	return &HTTPRetryClient{client: &http.Client{}, MaxAttempts: 1,
		shouldRetryOnStatus: shouldRetryOnStatus}
}

func defaultShouldRetryStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (c *HTTPRetryClient) Do(req *http.Request) (*http.Response, error) {

	var resp *http.Response
	var doErr error

	for attempt := 1; attempt <= c.MaxAttempts; attempt++ {
		// reset the request body
		if req.GetBody != nil {
			body, getBodyErr := req.GetBody()
			if getBodyErr != nil {
				return nil, getBodyErr
			}
			req.Body = body

		}
		resp, doErr = c.client.Do(req)
		if doErr == nil {
			shouldRetryStatus := c.shouldRetryOnStatus
			if shouldRetryStatus == nil {
				shouldRetryStatus = defaultShouldRetryStatus
			}

			if !shouldRetryStatus(resp.StatusCode) || attempt >= c.MaxAttempts {
				return resp, nil
			}

			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

		} else if attempt >= c.MaxAttempts {
			return resp, doErr
		}

		timer := time.NewTimer(c.AttemptDelay)
		select {
		case <-req.Context().Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, req.Context().Err()
		case <-timer.C:
		}

	}

	return resp, doErr
}
