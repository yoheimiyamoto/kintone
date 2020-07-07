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

func TestDateField(t *testing.T) {
	tests := []struct {
		Input    []byte
		Expected string
	}{
		{[]byte(`"2014-02-16"`), "2014-02-16"},
		{[]byte(`""`), ""},
	}

	for _, test := range tests {
		var d DateField
		err := json.Unmarshal(test.Input, &d)
		if err != nil {
			t.Error(err)
			return
		}

		actual := fmt.Sprint(d)
		if test.Expected != actual {
			t.Errorf("expected: %s, actual: %s", test.Expected, actual)
		}
	}
}

func TestDateTimeField(t *testing.T) {
	tests := []struct {
		Input    []byte
		Expected string
	}{
		{[]byte(`"2014-02-16T08:57:00Z"`), "2014-02-16 08:57:00"},
		{[]byte(`""`), ""},
	}

	for _, test := range tests {
		var d DateTimeField
		err := json.Unmarshal(test.Input, &d)
		if err != nil {
			t.Error(err)
			return
		}

		actual := fmt.Sprint(d)
		if test.Expected != actual {
			t.Errorf("expected: %s, actual: %s", test.Expected, actual)
		}
	}
}
