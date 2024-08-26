package downloader

import (
	"bytes"
	"io"
	"net/http"
)

type DownloaderClient struct {
	client *http.Client
}

func NewHttpClient() *DownloaderClient {
	return &DownloaderClient{
		client: &http.Client{},
	}
}

func (c *DownloaderClient) DoRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *DownloaderClient) NewRequest(method, url string, headers map[string]string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if body != nil {
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return req, nil
}

func (c *DownloaderClient) Do(method string, url string, headers map[string]string) (*http.Response, error) {
	req, err := c.NewRequest(method, url, headers, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
