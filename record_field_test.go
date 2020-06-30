package kintone

import (
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
