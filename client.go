package gj

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (c *Client) call(method, path string, body io.Reader) (int, []byte, error) {
	req, err := http.NewRequest(method, c.url+path, body)
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
	_, b, err := c.call("GET", "/api/v1/procs", nil)
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
	_, b, err := c.call("GET", fmt.Sprintf("/api/v1/procs/%s", pid), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to PS request: %s", err)
	}
	respModel := APIResponseShowProc{}
	if err := json.Unmarshal(b, &respModel); err != nil {
		return nil, fmt.Errorf("failed to parse json: %s", err)
	}
	return respModel.Proc, nil
}

func (c *Client) Create(r io.Reader) (string, error) {
	status, b, err := c.call("POST", "/api/v1/procs", r)
	if err != nil {
		return "", fmt.Errorf("failed to Create request: %s", err)
	}
	if status != 200 {
		return "", errors.New("failed to create process")
	}
	respModel := APIResponseCreateProc{}
	if err := json.Unmarshal(b, &respModel); err != nil {
		return "", fmt.Errorf("failed to parse json: %s", err)
	}
	return respModel.PID, nil
}

func (c *Client) Start(pid string) error {
	status, _, err := c.call("GET", fmt.Sprintf("/api/v1/procs/%s/start", pid), nil)
	if err != nil {
		return fmt.Errorf("failed to Start request: %s", err)
	}
	if status != 200 {
		return errors.New("start failed")
	}
	return nil
}

func (c *Client) Log(pid string) (string, error) {
	_, b, err := c.call("GET", fmt.Sprintf("/api/v1/procs/%s/log", pid), nil)
	if err != nil {
		return "", fmt.Errorf("failed to Log request: %s", err)
	}
	if err != nil {
		return "", fmt.Errorf("failed to read body: %s", err)
	}
	return string(b), nil
}
