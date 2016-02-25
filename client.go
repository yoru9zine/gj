package gj

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	*http.Client
	url string
}

func NewClient(url string) *Client {
	return &Client{
		Client: http.DefaultClient,
		url:    url,
	}
}

func (c *Client) call(method, path string) (int, []byte, error) {
	req, err := http.NewRequest(method, c.url+path, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %s", err)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to request api: %s", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read body: %s", err)
	}
	return resp.StatusCode, b, nil
}

func (c *Client) PS() (map[string]*ProcessViewModel, error) {
	_, b, err := c.call("GET", "/api/v1/procs")
	if err != nil {
		return nil, fmt.Errorf("failed to PS request: %s", err)
	}
	respModel := APIResponseShowProcs{}
	if err := json.Unmarshal(b, &respModel); err != nil {
		return nil, fmt.Errorf("failed to parse json: %s", err)
	}
	return respModel.Procs, nil
}

func (c *Client) Show(pid string) (*ProcessViewModel, error) {
	_, b, err := c.call("GET", fmt.Sprintf("/api/v1/procs/%s", pid))
	if err != nil {
		return nil, fmt.Errorf("failed to PS request: %s", err)
	}
	respModel := APIResponseShowProc{}
	if err := json.Unmarshal(b, &respModel); err != nil {
		return nil, fmt.Errorf("failed to parse json: %s", err)
	}
	return respModel.Proc, nil
}

func (c *Client) Start(pid string) error {
	status, _, err := c.call("GET", fmt.Sprintf("/api/v1/procs/%s/start", pid))
	if err != nil {
		return fmt.Errorf("failed to Start request: %s", err)
	}
	if status != 200 {
		return errors.New("start failed")
	}
	return nil
}

func (c *Client) Log(pid string) (string, error) {
	_, b, err := c.call("GET", fmt.Sprintf("/api/v1/procs/%s/log", pid))
	if err != nil {
		return "", fmt.Errorf("failed to Log request: %s", err)
	}
	if err != nil {
		return "", fmt.Errorf("failed to read body: %s", err)
	}
	return string(b), nil
}
