package kintone

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestUserField(t *testing.T) {
	u := &UserField{Code: "aaa", Name: "hello"}
	actual := fmt.Sprint(u)
	expected := "hello"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
		return
	}
}

func TestDateTimeField(t *testing.T) {
	data := []byte(`"2014-02-16T08:57:00Z"`)

	var d DateTimeField
	err := json.Unmarshal(data, &d)
	if err != nil {
		t.Error(err)
		return
	}

	expected := "2014-02-16 08:57:00"
	actual := fmt.Sprint(d)
	if expected != actual {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}
