package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mirror-sync/pkg/project"
	"mirror-sync/pkg/remote/obj"
	"net/http"
	"net/url"
)

type (
	Client struct {
		url string
	}

	SimpleError struct {
		Message string `json:"message"`
	}
)

func New(url string) *Client {
	return &Client{
		url: url,
	}
}

func (c *Client) Apply(pr project.Project) error {
	url, err := url.JoinPath(c.url, "api", "v1", "projects")
	if err != nil {
		return fmt.Errorf("failed to make url: %s", err)
	}

	data, err := json.Marshal(pr)
	if err != nil {
		return fmt.Errorf("failed to marshal project data: %s", err)
	}

	r := bytes.NewReader(data)

	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return fmt.Errorf("failed to generate http request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var cli http.Client
	res, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send the request to the server: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		return fmt.Errorf("failed to send the request to the server: %s: %s", res.Status, toError(res.Body))
	}

	return nil
}

func (c *Client) Version() (obj.SystemInformation, error) {
	url, err := url.JoinPath(c.url, "api", "v1", "version")
	if err != nil {
		return obj.SystemInformation{}, fmt.Errorf("failed to make url: %s", err)
	}

	res, err := http.Get(url)
	if err != nil {
		return obj.SystemInformation{}, fmt.Errorf("failed to send the request to the server: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return obj.SystemInformation{}, fmt.Errorf("failed to send the request to the server: %s: %s", res.Status, toError(res.Body))
	}

	var payload obj.HTTPObject[obj.SystemInformation]
	d := json.NewDecoder(res.Body)
	if err := d.Decode(&payload); err != nil {
		return obj.SystemInformation{}, fmt.Errorf("failed to parse the server response, is your client up-to-date? (reason: %s)", err)
	}

	return payload.Data, nil
}

func (c *Client) List() ([]project.Project, error) {
	url, err := url.JoinPath(c.url, "api", "v1", "projects", "all")
	if err != nil {
		return nil, fmt.Errorf("failed to make url: %s", err)
	}

	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send the request to the server: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to send the request to the server: %s: %s", res.Status, toError(res.Body))
	}

	var payload obj.HTTPObject[[]project.Project]
	d := json.NewDecoder(res.Body)
	if err := d.Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to parse the server response, is your client up-to-date? (reason: %s)", err)
	}

	prs := payload.Data

	for i, pr := range prs {
		pr.ServerURL = c.url
		prs[i] = pr
	}

	return prs, nil
}

func (c *Client) Remove(pr project.Project) error {
	url, err := url.JoinPath(c.url, "api", "v1", "projects")
	if err != nil {
		return fmt.Errorf("failed to make url: %s", err)
	}

	data, err := json.Marshal(pr)
	if err != nil {
		return fmt.Errorf("failed to marshal project data: %s", err)
	}

	r := bytes.NewReader(data)

	req, err := http.NewRequest("DELETE", url, r)
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err)
	}

	cli := http.Client{}
	res, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send the request to the server: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 204 {
		return fmt.Errorf("failed to send the request to the server: %s: %s", res.Status, toError(res.Body))
	}

	return nil
}

func (c *Client) RunOne(pr project.Project) error {
	url, err := url.JoinPath(c.url, "api", "v1", "run")
	if err != nil {
		return fmt.Errorf("failed to make url: %s", err)
	}

	data, err := json.Marshal(pr)
	if err != nil {
		return fmt.Errorf("failed to marshal project data: %s", err)
	}

	r := bytes.NewReader(data)

	req, err := http.NewRequest("EXECUTE", url, r)
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err)
	}

	cli := http.Client{}
	res, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send the request to the server: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("failed to send the request to the server: %s: %s", res.Status, toError(res.Body))
	}

	return nil
}

func toError(body io.ReadCloser) error {
	var msg SimpleError

	d := json.NewDecoder(body)
	if err := d.Decode(&msg); err != nil {
		return fmt.Errorf("failed to decode error message: %s", err)
	}

	return fmt.Errorf("%s", msg.Message)
}
