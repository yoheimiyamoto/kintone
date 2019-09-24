package kintone

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// APIEndpoint constants
const (
	APIEndpointBase       = "https://%s.cybozu.com"
	APIEndpointRecord     = "/k/v1/record.json"
	APIEndpointRecords    = "/k/v1/records.json"
	APIEndpointApp        = "/k/v1/app.json"
	APIEndpointFormField  = "/k/v1/app/form/fields.json"
	APIEndpointFormLayout = "/k/v1/app/form/layout.json"
	APIEndpointFile       = "/k/v1/file.json"
)

// Client ...
type Client interface {
	get(path string, q *Query) ([]byte, error)
	post(path string, body []byte) ([]byte, error)
	put(path string, body []byte) ([]byte, error)
	delete(path string, body []byte) ([]byte, error)
}

// Client ...
type client struct {
	username     string
	password     string
	apiToken     string
	endpointBase *url.URL
	httpClient   *http.Client
}

// DefaultTimeout ...
const (
	DefaultTimeout = time.Second * 600 // Default value for App.Timeout
)

// Library internal errors.
var (
	ErrTimeout         = errors.New("Timeout")
	ErrInvalidResponse = errors.New("Invalid Response")
	ErrTooMany         = errors.New("Too many records")
)

// ClientError ...
type ClientError struct {
	HTTPStatus     string `json:"-"`       // e.g. "404 NotFound"
	HTTPStatusCode int    `json:"-"`       // e.g. 404
	Message        string `json:"message"` // Human readable message.
	ID             string `json:"id"`      // A unique error ID.
	Code           string `json:"code"`    // For machines.
	Errors         string `json:"errors"`  // Error Description.
}

func (e *ClientError) Error() string {
	if e.Message == "" {
		return "HTTP error: " + e.HTTPStatus
	}
	return fmt.Sprintf("AppError: %d [%s] %s (%s) %s",
		e.HTTPStatusCode, e.Code, e.Message, e.ID, e.Errors)
}

// NewClient ...
func newClient(subdomain string, username, password string, httpClient *http.Client) *client {
	c := client{
		username: username,
		password: password,
	}

	if httpClient != nil {
		c.httpClient = httpClient
	} else {
		c.httpClient = &http.Client{Timeout: DefaultTimeout}
	}

	u, _ := url.ParseRequestURI(fmt.Sprintf(APIEndpointBase, subdomain))
	c.endpointBase = u

	c.apiToken = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	return &c
}

func (c *client) get(path string, q *Query) ([]byte, error) {
	url, err := newURL(c.endpointBase, path, q)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return c.do(req)
}

func (c *client) post(path string, body []byte) ([]byte, error) {
	url, err := newURL(c.endpointBase, path, nil)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	return c.do(req)
}

func (c *client) put(path string, body []byte) ([]byte, error) {
	url, err := newURL(c.endpointBase, path, nil)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	return c.do(req)
}

func (c *client) delete(path string, body []byte) ([]byte, error) {
	url, err := newURL(c.endpointBase, path, nil)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	return c.do(req)
}

func (c *client) do(req *http.Request) ([]byte, error) {
	req.Header.Set("X-Cybozu-Authorization", c.apiToken)
	req.SetBasicAuth(c.username, c.password)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		var e resError
		err = json.Unmarshal(body, &e)
		if err != nil {
			return nil, err
		}
		return nil, &e
	}

	return body, nil
}

func newURL(endpointBase *url.URL, path string, q *Query) (string, error) {
	u := *endpointBase
	u.Path = path

	if q != nil {
		q, err := url.ParseQuery(q.String())
		if err != nil {
			return "", err
		}
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}