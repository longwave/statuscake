package statuscake

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const apiBaseURL = "https://www.statuscake.com/API"

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type apiClient interface {
	get(string) (*http.Response, error)
	delete(string, url.Values) (*http.Response, error)
	put(string, url.Values) (*http.Response, error)
}

// Client is the http client that wraps the remote API.
type Client struct {
	c           httpClient
	username    string
	apiKey      string
	testsClient Tests
}

// New returns a new Client
func New(username string, apiKey string) *Client {
	c := &http.Client{}
	return &Client{
		c:        c,
		username: username,
		apiKey:   apiKey,
	}
}

func (c *Client) newRequest(method string, path string, v url.Values, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", apiBaseURL, path)
	if v != nil {
		url = fmt.Sprintf("%s?%s", url, v.Encode())
	}

	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Username", c.username)
	r.Header.Set("API", c.apiKey)

	return r, nil
}

func (c *Client) doRequest(r *http.Request) (*http.Response, error) {
	resp, err := c.c.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &httpError{
			status:     resp.Status,
			statusCode: resp.StatusCode,
		}
	}

	return resp, nil
}

func (c *Client) get(path string) (*http.Response, error) {
	r, err := c.newRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	return c.doRequest(r)
}

func (c *Client) put(path string, v url.Values) (*http.Response, error) {
	r, err := c.newRequest("PUT", path, nil, strings.NewReader(v.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		return nil, err
	}

	return c.doRequest(r)
}

func (c *Client) delete(path string, v url.Values) (*http.Response, error) {
	r, err := c.newRequest("DELETE", path, v, nil)
	if err != nil {
		return nil, err
	}

	return c.doRequest(r)
}

// Tests returns a client that implements the `Tests` API.
func (c *Client) Tests() Tests {
	if c.testsClient == nil {
		c.testsClient = newTests(c)
	}

	return c.testsClient
}
