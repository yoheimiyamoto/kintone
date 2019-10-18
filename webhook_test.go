package kintone

import (
	"encoding/json"
	"testing"
)

func TestWebhook(t *testing.T) {
	data := []byte(`
		{
			"id":"01234567-0123-0123-0123-0123456789ab",
			"type":"ADD_RECORD",
			"app":{
				"id":"1",
				"name":"案件管理"
			},
			"record":{
				"レコード番号":{
					"type":"RECORD_NUMBER",
					"value":"2"
				}
			},
			"recordTitle":"往訪：サイボウズ株式会社",
			"url":"https://example.cybozu.com/k/1/show#record=2"
		}
	`)
	var webhook Webhook

	err := json.Unmarshal(data, &webhook)
	if err != nil {
		t.Error(err)
		return
	}

	expected := "rpy"
	actual := webhook.URL.Subdomain()

	if expected != actual {
		t.Logf("expected: %s, actual: %s", expected, actual)
		return
	}
}

func TestURL(t *testing.T) {
	fn := func(input, expected string) {
		t.Helper()
		url := URL(input)
		actual := url.Subdomain()
		if expected != actual {
			t.Errorf("expected: %s, actual: %s", expected, actual)
			return
		}
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"https://rpy.cybozu.com", "rpy"},
		{"https://cybozu.com", ""},
	}

	for _, test := range tests {
		fn(test.input, test.expected)
	}
}
