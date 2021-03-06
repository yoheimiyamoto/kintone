package kintone

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
)

func TestRead(t *testing.T) {
	op := RepositoryOption{MaxRetry: 3}
	// repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &op)
	// repo.Client.SetBasicAuth(os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"))

	repo := NewRepository("rls-shinsa", "GCPkintone", "shinsa12", &op)
	repo.Client.SetBasicAuth("Administrator", "Admin12")

	// q := &Query{AppID: 1002, Condition: `value="upsert via cloud functions!!"`, Fields: []string{"レコード番号"}}
	q := &Query{AppID: 91, Condition: `受付番号="123"`, Fields: []string{"レコード番号"}}

	rs, err := repo.ReadRecords(context.Background(), q)
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

func TestQuery(t *testing.T) {
	q := Query{ID: 1}
	actual := q.String()
	expected := "id=1"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
		return
	}
	t.Log(q.String())
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

func TestAddRecord(t *testing.T) {
	op := RepositoryOption{MaxRetry: 3}
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &op)
	r := Record{Fields: Fields{"キー": SingleLineTextField("test")}}
	id, err := repo.AddRecord(nil, 688, &r)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(id)
}

func TestAddRecords(t *testing.T) {
	op := RepositoryOption{MaxRetry: 3}
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &op)
	var rs []*Record
	for i := 0; i < 2; i++ {
		rs = append(rs, &Record{Fields: Fields{
			"グループ名":     SingleLineTextField(fmt.Sprintf("hello%d", i)),
			"文字列__複数行_": MultiLineTextField("hello world!"),
			"チェックボックス":  CheckBoxField([]string{"sample1"}),
			"ドロップダウン":   NewSingleSelectField("sample1"),
			"複数選択":      MultiSelectField([]string{"sample1", "sample2"}),
			"ユーザー選択":    []*UserField{&UserField{Code: "yoheimiyamoto"}},
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

func TestAddWithRetry(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	var rs []*Record
	for i := 0; i < 1; i++ {
		rs = append(rs, &Record{Fields: Fields{
			"グループ名":     SingleLineTextField(fmt.Sprintf("hello%d", i)),
			"文字列__複数行_": MultiLineTextField("hello world!"),
			"チェックボックス":  CheckBoxField([]string{"sample1"}),
			"ドロップダウン":   NewSingleSelectField("sample1"),
			"複数選択":      MultiSelectField([]string{"sample1", "sample2"}),
			"ユーザー選択":    []*UserField{{Code: "yoheimiyamoto"}},
			"日付":        NewDateField(2011, 1, 1),
			"テーブル": TableField([]*Record{{Fields: Fields{
				"テーブルフィールド1": SingleLineTextField("hello"),
				"テーブルフィールド2": SingleLineTextField("hello"),
			}}}),
		}})
	}
	ids, err := repo.addRecordsWithRetry(context.Background(), 688, rs)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(ids)
}

func TestUpdateRecord(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	// r := Record{Fields: Fields{"キー": SingleLineTextField("test"), "数値": NumberField(1)}}
	r := Record{ID: "6173", Fields: Fields{"キー": SingleLineTextField("test"), "数値": NumberField(100)}}

	err := repo.UpdateRecord(nil, 688, "", &r)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestUpdateRecords(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	rs := []*Record{
		{ID: "1662", Fields: Fields{"name": SingleLineTextField("卍<>*+!~^")}},
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

func TestReadSpace(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)
	s, err := repo.ReadSpace(7)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", s)
}

func TestAddSpace(t *testing.T) {
	repo := NewRepository("rcl-trial", os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), nil)

	s := CreateSpace{
		TemplateID: 1,
		Name:       "test",
		Members: []*CreateSpaceMember{
			&CreateSpaceMember{
				EntityType: "USER",
				Code:       "test@gmail.com",
				IsAdmin:    true,
			},
		},
		IsPrivate: true,
	}

	id, err := repo.AddSpace(&s)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(id)
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

func TestReadRecordsWithCursor(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &RepositoryOption{MaxConcurrent: 90})

	q := NewQuery(1002)
	q.Condition = `レコード番号="212002"`
	q.Fields = []string{"name"}

	rs, err := repo.ReadRecordsWithCursor(q)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(rs[0])
}

func TestGetCursor(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &RepositoryOption{MaxConcurrent: 90})

	q := NewQuery(1002)

	c, err := repo.getCursor(q)
	if err != nil {
		t.Error(err)
	}

	t.Log(c)
}

func TestUpsertRecord(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &RepositoryOption{MaxConcurrent: 5})
	// r := Record{
	// 	ID:     "1",
	// 	Fields: Fields{"キー": SingleLineTextField("test"), "数値": NumberField(1)},
	// }
	r := Record{
		ID:     "6173000",
		Fields: Fields{"数値": NumberField(10)},
	}

	id, err := repo.UpsertRecord(nil, 688, "", &r)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(id)
}

func TestUpsertRecords(t *testing.T) {
	repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &RepositoryOption{MaxConcurrent: 5})

	start := 200388
	length := 1000
	var rs []*Record

	for i := start; i < start+length; i++ {
		id := fmt.Sprint(i)
		rs = append(rs, &Record{ID: id, Fields: Fields{
			"id":   SingleLineTextField(id),
			"name": SingleLineTextField("hello"),
		}})
	}

	err := repo.UpsertRecords(context.Background(), 1001, "id", rs...)
	if err != nil {
		t.Error(err)
	}
}
