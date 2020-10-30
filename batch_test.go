package kintone

import (
	"fmt"
	"os"
	"testing"
)

func TestAddSpaces(t *testing.T) {
	repo := NewRepository("rcl-trial", os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	// repo := NewRepository("rcl-trial", "yoheimiyamoto.work@gmail.com", "98wtla50", nil)

	fn := func(mail string) (int, error) {
		s := CreateSpace{
			TemplateID: 2,
			Name:       "マイスペース",
			Members: []*CreateSpaceMember{
				&CreateSpaceMember{
					EntityType: "USER",
					Code:       mail,
					IsAdmin:    true,
				},
				&CreateSpaceMember{
					EntityType: "USER",
					Code:       "Administrator",
					IsAdmin:    true,
				},
			},
			IsPrivate: true,
		}

		return repo.AddSpace(&s)
	}

	mails := []string{
		"test@gmail.com",
	}

	for _, m := range mails {
		id, err := fn(m)
		if err != nil {
			t.Errorf("%s failed: err: %v", m, err)
			return
		}
		fmt.Printf("created: mail: %s, spaceID: %d\n", m, id)
	}
}
