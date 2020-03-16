package kintone

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
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
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	// q := &Query{AppID: 1002, Condition: `value="upsert via cloud functions!!"`, Fields: []string{"レコード番号"}}
	q := &Query{AppID: 1002, Condition: `value="upsert via cloud functions 5"`, Fields: []string{"レコード番号"}}
	rs, err := repo.readRecords(context.Background(), q)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(len(rs))
}

func TestReadWithOrderBy(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	q := &Query{AppID: 688, OrderBy: "グループ名 desc", Fields: []string{"name", "グループ名", "日付"}}
	rs, err := repo.ReadRecords(nil, q)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(len(rs))
	t.Logf("%v", rs[0])
}

func TestReadWithQuery1(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &RepositoryOption{MaxConcurrent: 5})

	condition := `value="upsert via cloud functions!!"`

	q := &Query{AppID: 1002, Fields: []string{"レコード番号"}, Condition: condition}
	rs, err := repo.ReadRecords(nil, q)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%d records", len(rs))
}

func TestReadWithQuery2(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)

	start := 20002
	length := 100

	var ids []string
	for i := start; i < start+length; i++ {
		ids = append(ids, strconv.Itoa(i))
	}

	condition := ""

	for i, id := range ids {
		if i == 0 {
			condition = fmt.Sprintf(`レコード番号="%s"`, id)
			continue
		}

		condition += fmt.Sprintf(` or レコード番号="%s"`, id)
	}

	q := &Query{AppID: 1002, Fields: []string{"レコード番号"}, Condition: condition}
	rs, err := repo.ReadRecords(nil, q)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%d records", len(rs))
}

func TestRead500Records(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)

	start := 20002
	length := 500

	var ids []string
	for i := start; i < start+length; i++ {
		ids = append(ids, strconv.Itoa(i))
	}
	log.Printf("%d ids", len(ids))

	condition := ""

	for i, id := range ids {
		if i == 0 {
			condition = fmt.Sprintf(`レコード番号="%s"`, id)
			continue
		}

		condition += fmt.Sprintf(` or レコード番号="%s"`, id)
	}

	q := &Query{AppID: 1002, Fields: []string{"レコード番号"}, Condition: condition}
	rs, err := repo.readRecords(context.Background(), q)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%d records", len(rs))
}
func TestAdd(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	var rs []*Record
	for i := 0; i < 1; i++ {
		rs = append(rs, &Record{Fields: Fields{
			"グループ名":     SingleLineTextField(fmt.Sprintf("hello%d", i)),
			"文字列__複数行_": MultiLineTextField("hello world!"),
			"チェックボックス":  CheckBoxField([]string{"sample1"}),
			"ドロップダウン":   NewSingleSelectField("sample1"),
			"複数選択":      MultiSelectField([]string{"sample1", "sample2"}),
			"ユーザー選択":    UsersField{&CodeField{Code: "yoheimiyamoto"}},
			"日付":        NewDateField(2011, 1, 1),
			"テーブル": TableField([]*Record{&Record{Fields: Fields{
				"テーブルフィールド1": SingleLineTextField("hello"),
				"テーブルフィールド2": SingleLineTextField("hello"),
			}}}),
		}})
	}
	ids, err := repo.AddRecords(nil, 688, rs...)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(ids)
}

func TestUpdate(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	rs := []*Record{
		&Record{ID: "1662", Fields: Fields{"name": SingleLineTextField("卍<>*+!~^")}},
	}
	err := repo.UpdateRecords(nil, 670, "", rs...)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestReadAndUpdate(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)

	q := &Query{AppID: 688, Condition: `レコード番号="3145"`}

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

	err = repo.UpdateRecords(nil, 688, "", rs...)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestUpdateWithKey(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	rs := []*Record{&Record{ID: "1662", Fields: Fields{"id": SingleLineTextField("100"), "name": SingleLineTextField("world!!")}}}
	err := repo.UpdateRecords(nil, 670, "id", rs...)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestReadFormFields(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	fs, err := repo.ReadFormFields(688)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", fs["チェックボックス"].Options[0])
}

func TestBulkAdds(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &RepositoryOption{MaxConcurrent: 90})
	var rs []*Record

	for i := 0; i < 1000; i++ {
		code := fmt.Sprintf("code_%d", i)
		fs := Fields{
			"グループ名": SingleLineTextField(code),
		}
		rs = append(rs, &Record{Fields: fs})
	}

	_, err := repo.AddRecords(nil, 688, rs...)
	if err != nil {
		t.Error(err)
	}
}

func TestUpsertRecords(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &RepositoryOption{MaxConcurrent: 5})

	length := 1000
	rs := make([]*Record, length)

	for i := 0; i < length; i++ {
		id := fmt.Sprint(i)
		rs[i] = &Record{ID: id, Fields: Fields{
			"id":        SingleLineTextField(id),
			"upsert_id": SingleLineTextField(id),
			"value":     SingleLineTextField("cloud functions test 7"),
		}}
	}

	err := repo.UpsertRecords(context.Background(), 1002, "upsert_id", rs...)
	if err != nil {
		t.Error(err)
	}
}

func TestUpsertKadenRecords(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &RepositoryOption{MaxConcurrent: 5})

	//+ids
	rs, err := repo.ReadRecords(ctx, &Query{AppID: 664, Condition: "", Fields: []string{"VM加盟店番号_関連レコード用"}})
	if err != nil {
		t.Error(err)
		return
	}
	ids := make([]string, len(rs))
	for i, r := range rs {
		ids[i] = fmt.Sprint(r.Fields["VM加盟店番号_関連レコード用"])
	}

	ids = ids[0:1000]

	log.Printf("%d ids", len(ids))
	//-ids

	rs = make([]*Record, len(ids))

	for i, id := range ids {
		rs[i] = &Record{ID: id, Fields: Fields{
			"VM加盟店番号_関連レコード用":       SingleLineTextField(id),
			"APPLY_DECIDING_FACTOR": SingleLineTextField("test4"),
		}}
	}

	err = repo.UpsertRecords(context.Background(), 664, "VM加盟店番号_関連レコード用", rs...)
	if err != nil {
		t.Error(err)
	}
}
