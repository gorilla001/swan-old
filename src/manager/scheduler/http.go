package scheduler

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	HttpTimeoutDuration   = 10 * time.Second
	HttpKeepaliveDuration = 30 * time.Second

	UserAgent = "swan"

	MesosSchedulerAPI = "/api/v1/scheduler"
)

type httpClient struct {
	streamID string
	endPoint string
	client   *http.Client
}

func NewHTTPClient(leader string) *HttpClient {
	return &httpClient{
		endPoint: "http://" + leader + MesosSchedulerAPI,
		client: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   HttpTimeoutDuration,
					KeepAlive: HttpKeepaliveDuration,
				}).Dial,
			},
		},
	}
}

func (c *HttpClient) send(payload []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest("POST", c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", UserAgent)
	if c.streamID != "" {
		httpReq.Header.Set("Mesos-Stream-Id", c.streamID)
	}

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request: %s", err)
	}

	if httpResp.Header.Get("Mesos-Stream-Id") != "" {
		c.streamID = httpResp.Header.Get("Mesos-Stream-Id")
	}

	return httpResp, nil
}
