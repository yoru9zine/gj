package gj

import (
	"encoding/json"
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

func (c *Client) PS() (map[string]*JobViewModel, error) {
	req, err := http.NewRequest("GET", c.url+"/api/v1/jobs", nil)
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
	respModel := APIResponseShowJobs{}
	if err := json.Unmarshal(b, &respModel); err != nil {
		return nil, fmt.Errorf("failed to parse json: %s", err)
	}
	return respModel.Jobs, nil
}
