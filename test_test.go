package kintone

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load("kintone.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func TestJsonEqual(t *testing.T) {
	data1 := []byte(`{"name":"hello", "age":100}`)
	data2 := []byte(`{"age":100, "name":"hello"}`)

	t.Log(jsonEqual(data1, data2))
}

// jsonの比較
func jsonEqual(expected, actual []byte) bool {
	var d1 interface{}
	json.Unmarshal(expected, &d1)

	var d2 interface{}
	json.Unmarshal(actual, &d2)

	return reflect.DeepEqual(d1, d2)
}
