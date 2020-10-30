package kintone

import (
	"fmt"
	"net/url"
	"testing"
)

type MockClient struct{}

func (c *MockClient) get(path string, q *Query) ([]byte, error) {
	b := []byte(`{}`)
	return b, nil
}

func (c *MockClient) post(path string, q *Query) ([]byte, error) {
	return nil, nil
}

func (c *MockClient) put(path string, q *Query) ([]byte, error) {
	return nil, nil
}

func (c *MockClient) delete(path string, q *Query) ([]byte, error) {
	return nil, nil
}

func TestNewURL(t *testing.T) {
	endPointBase, _ := url.ParseRequestURI(fmt.Sprintf(APIEndpointBase, "rpy"))
	endPointBase.RawQuery = "id=1"
	u, err := newURL(endPointBase, "hello", nil)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(u)
}
