package kintone

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
