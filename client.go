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

func (c *Client) PS() (map[string]*ProcessViewModel, error) {
	req, err := http.NewRequest("GET", c.url+"/api/v1/procs", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request api: %s", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %s", err)
	}
	respModel := APIResponseShowProcs{}
	if err := json.Unmarshal(b, &respModel); err != nil {
		return nil, fmt.Errorf("failed to parse json: %s", err)
	}
	return respModel.Procs, nil
}

func (c *Client) Start(pid string) error {
	req, err := http.NewRequest("GET", c.url+fmt.Sprintf("/api/v1/procs/%s/start", pid), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request api: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("start failed")
	}
	return nil
}

func (c *Client) Log(pid string) (string, error) {
	req, err := http.NewRequest("GET", c.url+fmt.Sprintf("/api/v1/procs/%s/log", pid), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %s", err)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request api: %s", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %s", err)
	}
	return string(b), nil
}
