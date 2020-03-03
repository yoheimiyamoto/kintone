package kintone

import (
	"context"
	"os"
	"strconv"
	"testing"
)

func TestBalk(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &RepositoryOption{MaxConcurrent: 10})
	var rs []*Record
	for i := 2; i < 200000; i++ {
		rs = append(rs, &Record{ID: strconv.Itoa(i), Fields: Fields{
			"value": SingleLineTextField("upsert value2"),
		}})
	}
	err := repo.UpsertRecords(context.Background(), 1002, "", rs...)
	if err != nil {
		t.Error(err)
		return
	}
	// t.Log(ids)
}
