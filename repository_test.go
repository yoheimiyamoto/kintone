package kintone

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load("kintone.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func TestRead(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil, 1)
	q := &Query{AppID: "670", RawQuery: &RawQuery{Condition: `name="world"`}, Fields: []string{"name", "日付"}}
	// q := &Query{AppID: 670, RawQuery: &RawQuery{Condition: ``}, Fields: []string{"name", "日付"}}
	// q := &Query{AppID: 670, RawQuery: &RawQuery{Condition: ``}}

	rs, err := repo.ReadRecords(nil, q)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(len(rs))
}

func TestAdd(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil, 1)
	var rs []*Record
	for i := 0; i < 1; i++ {
		rs = append(rs, &Record{Fields: Fields{
			"グループ名":     SingleLineTextField(fmt.Sprintf("hello%d", i)),
			"文字列__複数行_": MultiLineTextField("hello world!"),
			"チェックボックス":  CheckBoxField([]string{"sample1"}),
			"ドロップダウン":   SingleSelectField("sample1"),
			"複数選択":      MultiSelectField([]string{"sample1", "sample2"}),
			"ユーザー選択":    UsersField{&CodeField{Code: "yoheimiyamoto"}},
			"日付":        NewDateField(2011, 1, 1),
			"テーブル": TableField([]*Record{&Record{Fields: Fields{
				"テーブルフィールド1": SingleLineTextField("hello"),
				"テーブルフィールド2": SingleLineTextField("hello"),
			}}}),
		}})
	}
	ids, err := repo.AddRecords(nil, "688", rs)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(ids)
}

func TestUpdate(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil, 1)
	rs := []*Record{&Record{ID: "1662", Fields: Fields{"name": SingleLineTextField("world")}}}
	err := repo.UpdateRecords(nil, "670", "", rs...)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestReadAndUpdate(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil, 1)

	q := &Query{AppID: "688", RawQuery: &RawQuery{Condition: `レコード番号="3145"`}}

	rs, err := repo.ReadRecords(nil, q)
	if err != nil {
		t.Error(err)
		return
	}

	r := rs[0]
	table := rs[0].Fields["テーブル"]
	if v, ok := table.(TableField); ok {
		v[0].Fields["テーブルフィールド1"] = SingleLineTextField("hello!!!!")
	}

	r.Fields = Fields{"テーブル": table}

	rs = []*Record{r}

	err = repo.UpdateRecords(nil, "688", "", rs...)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestUpdateWithKey(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil, 1)
	rs := []*Record{&Record{ID: "1662", Fields: Fields{"id": SingleLineTextField("100"), "name": SingleLineTextField("world!!")}}}
	err := repo.UpdateRecords(nil, "670", "id", rs...)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestReadFormFields(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil, 1)
	fs, err := repo.ReadFormFields("688")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", fs["チェックボックス"].Options[0])
}

func TestBulkAdds(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil, 90)
	var rs []*Record

	for i := 1000000; i < 2000000; i++ {
		code := fmt.Sprintf("code_%d", i)
		name := fmt.Sprintf("name_%d", i)
		fs := Fields{
			"company_cd":        SingleLineTextField(code),
			"company_name":      SingleLineTextField(name),
			"company_name_kana": SingleLineTextField("hello"),
			"tel":               SingleLineTextField("03-1111-111"),
			"jigyo":             SingleLineTextField("hello"),
			"tantou":            SingleLineTextField("hello"),
			"tantou_kana":       SingleLineTextField("hello"),
			"temp_regist":       CheckBoxField([]string{"仮登録"}),
		}
		rs = append(rs, &Record{Fields: fs})
	}

	_, err := repo.AddRecords(nil, "95", rs)
	if err != nil {
		t.Error(err)
	}
}
