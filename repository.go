package kintone

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	RetryInterval = 10 // second
)

// Repository ...
type Repository struct {
	Client   Client
	Token    chan struct{}
	MaxRetry int
}

type RepositoryOption struct {
	HTTPClient    *http.Client
	MaxConcurrent int
	MaxRetry      int
}

type Cursor struct {
	ID         string `json:"id"`
	TotalCount int    `json:"totalCount,string"`
}

// NewRepository ...
func NewRepository(subdomain string, username, password string, option *RepositoryOption) *Repository {
	var httpClient *http.Client
	maxConcurrent := 1
	maxRetry := 3

	if option != nil {
		if option.HTTPClient != nil {
			httpClient = option.HTTPClient
		}

		if option.MaxConcurrent != 0 {
			maxConcurrent = option.MaxConcurrent
		}

		if option.MaxRetry != 0 {
			maxRetry = option.MaxRetry
		}
	}

	c := newClient(subdomain, username, password, httpClient)
	token := make(chan struct{}, maxConcurrent)
	u, _ := url.ParseRequestURI(fmt.Sprintf(APIEndpointBase, subdomain))
	c.endpointBase = u
	return &Repository{c, token, maxRetry}
}

// ReadRecords ...
func (repo *Repository) ReadRecords(ctx context.Context, q *Query) ([]*Record, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// レコード数確認
	totalCount, err := repo.readTotalCount(q)
	if err != nil {
		return nil, errors.Wrap(err, "read total count failed")
	}

	recordsCh := make(chan *Record, totalCount)

	q.limit = 500

	eg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < totalCount; i += 500 {

		// クエリ生成（コピー）
		q := *q
		q.offset = i
		eg.Go(func() error {
			_rs, err := repo.readRecords(ctx, &q)
			if err != nil {
				return err
			}
			for _, r := range _rs {
				recordsCh <- r
			}
			return nil
		})
	}

	// closer
	err = eg.Wait()
	if err != nil {
		return nil, err
	}
	close(recordsCh)

	// readRecords from recordsCh
	var rs []*Record
	for r := range recordsCh {
		rs = append(rs, r)
	}

	return rs, nil
}

// read 500 records
func (repo *Repository) readRecords(ctx context.Context, q *Query) ([]*Record, error) {
	q.limit = 500

	select {
	case repo.Token <- struct{}{}:
		defer func() {
			<-repo.Token
		}()
	case <-ctx.Done(): // cancelled
		return nil, errors.New("canceled")
	}

	body, err := repo.Client.get(APIEndpointRecords, q)
	if err != nil {
		return nil, err
	}

	r := struct {
		Records []*Record `json:"records"`
	}{}

	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	return r.Records, nil
}

func (repo *Repository) readTotalCount(q *Query) (int, error) {
	q.TotalCount = true

	body, err := repo.Client.get(APIEndpointRecords, q)
	if err != nil {
		return 0, err
	}

	r := struct {
		TotalCount int `json:"totalCount,string"`
	}{}

	if err := json.Unmarshal(body, &r); err != nil {
		return 0, err
	}

	return r.TotalCount, nil
}

func (repo *Repository) ReadRecordsWithCursor(q *Query) ([]*Record, error) {
	c, err := repo.getCursor(q)
	if err != nil {
		return nil, errors.Wrap(err, "get cursor failed")
	}

	var resuestBody = struct {
		ID string `json:"id"`
	}{c.ID}

	var rs []*Record

	for {
		body, err := json.Marshal(resuestBody)
		if err != nil {
			return nil, err
		}

		body, err = repo.Client.getWithBody(APIEndpointRecordsCursor, body)
		if err != nil {
			return nil, err
		}

		var response struct {
			Records []*Record `json:"records"`
			Next    bool      `json:"next"`
		}

		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}

		rs = append(rs, response.Records...)

		if !response.Next {
			return rs, nil
		}
	}
}

func (repo *Repository) getCursor(q *Query) (*Cursor, error) {
	if q == nil {
		return nil, errors.New("query is required")
	}

	type requestBody struct {
		AppID    int      `json:"app"`
		Fields   []string `json:"fields"`
		Conditon string   `json:"query"`
		Size     int      `json:"size"`
	}

	request := requestBody{AppID: q.AppID, Fields: q.Fields, Conditon: q.Condition, Size: 500}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	body, err = repo.Client.post(APIEndpointRecordsCursor, body)
	if err != nil {
		return nil, err
	}

	var cursor Cursor

	err = json.Unmarshal(body, &cursor)
	if err != nil {
		return nil, err
	}

	return &cursor, nil
}

//+AddRecord

// AddRecords ...
func (repo *Repository) AddRecords(ctx context.Context, appID int, rs ...*Record) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	sliced := sliceRecords(rs, 100)

	var ids []string

	eg, ctx := errgroup.WithContext(ctx)
	for _, _rs := range sliced {
		_rs := _rs
		eg.Go(func() error {
			return func() error {
				_ids, err := repo.addRecordsWithRetry(ctx, appID, _rs)
				if err != nil {
					return err
				}
				ids = append(ids, _ids...)
				return nil
			}()
		})
	}

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (repo *Repository) AddRecord(ctx context.Context, appID int, r *Record) (string, error) {
	type requestBody struct {
		App    int    `json:"app"`
		Record Fields `json:"record"`
	}

	body, err := json.Marshal(requestBody{appID, r.Fields})
	if err != nil {
		return "", err
	}

	body, err = repo.Client.post(APIEndpointRecord, body)
	if err != nil {
		return "", err
	}

	var reponseBody struct {
		ID string `json:"id"`
	}

	err = json.Unmarshal(body, &reponseBody)
	if err != nil {
		return "", err
	}

	return reponseBody.ID, nil
}

func (repo *Repository) addRecords(ctx context.Context, appID int, rs []*Record) ([]string, error) {
	if len(rs) == 0 {
		return nil, nil
	}

	select {
	case repo.Token <- struct{}{}: // acquire token
		defer func() {
			<-repo.Token
		}()
	case <-ctx.Done(): // cancelled
		return nil, errors.New("canceled")
	}

	type requestBody struct {
		App     int      `json:"app"`
		Records []Fields `json:"records"`
	}

	fs := make([]Fields, len(rs))
	for i, r := range rs {
		fs[i] = r.Fields
	}

	body, err := json.Marshal(requestBody{appID, fs})
	if err != nil {
		return nil, err
	}

	body, err = repo.Client.post(APIEndpointRecords, body)
	if err != nil {
		return nil, err
	}

	var reponseBody struct {
		IDs []string `json:"ids"`
	}

	err = json.Unmarshal(body, &reponseBody)
	if err != nil {
		return nil, err
	}

	return reponseBody.IDs, nil
}

func (repo *Repository) addRecordsWithRetry(ctx context.Context, appID int, rs []*Record) ([]string, error) {
	if len(rs) == 0 {
		return nil, nil
	}

	select {
	case repo.Token <- struct{}{}: // acquire token
		defer func() {
			<-repo.Token
		}()
	case <-ctx.Done(): // cancelled
		return nil, errors.New("canceled")
	}

	type requestBody struct {
		App     int      `json:"app"`
		Records []Fields `json:"records"`
	}

	fs := make([]Fields, len(rs))
	for i, r := range rs {
		fs[i] = r.Fields
	}

	var retryCount int

	requestData, err := json.Marshal(requestBody{appID, fs})
	if err != nil {
		return nil, err
	}

	var resData []byte

	for {
		resData, err = repo.Client.post(APIEndpointRecords, requestData)
		if err == nil {
			break
		}

		retryCount++
		if retryCount > repo.MaxRetry {
			break
		}
		log.Printf("retry: body: %s, err: %s", string(requestData), err.Error())
		time.Sleep(time.Second * RetryInterval)
	}

	if err != nil {
		return nil, err
	}

	var reponseBody struct {
		IDs []string `json:"ids"`
	}

	err = json.Unmarshal(resData, &reponseBody)
	if err != nil {
		return nil, err
	}

	return reponseBody.IDs, nil
}

//-AddRecord

//+UpdateRecord
func (repo *Repository) UpdateRecord(ctx context.Context, appID int, updateKey string, r *Record) error {
	if appID == 0 {
		return errors.New("appID is required")
	}

	if r == nil {
		return nil
	}

	type RequestBody interface{}

	type RequestBodyWithRecordID struct {
		App    int    `json:"app"`
		ID     string `json:"id"`
		Fields Fields `json:"record"`
	}

	type UpdateKey struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}

	type RequestBodyWithUpdateKey struct {
		App       int        `json:"app"`
		UpdateKey *UpdateKey `json:"updateKey"`
		Fields    Fields     `json:"record"`
	}

	var requestBody RequestBody

	if updateKey == "" {
		requestBody = &RequestBodyWithRecordID{App: appID, ID: r.ID, Fields: r.Fields}
	} else {
		u := UpdateKey{Field: updateKey, Value: fmt.Sprint(r.Fields[updateKey])}
		delete(r.Fields, updateKey)
		requestBody = &RequestBodyWithUpdateKey{App: appID, UpdateKey: &u, Fields: r.Fields}
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	_, err = repo.Client.put(APIEndpointRecord, body)
	if err != nil {
		return err
	}

	return nil
}

//-UpdateRecord

//+UpdateRecords

// UpdateRecords ...
func (repo *Repository) UpdateRecords(ctx context.Context, appID int, updateKey string, rs ...*Record) error {
	if ctx == nil {
		ctx = context.Background()
	}

	sliced := sliceRecords(rs, 100)

	eg, ctx := errgroup.WithContext(ctx)
	for _, _rs := range sliced {
		_rs := _rs
		eg.Go(func() error {
			return func() error {
				err := repo.updateRecordsWithRetry(ctx, appID, _rs, updateKey)
				if err != nil {
					return err
				}
				return nil
			}()
		})
	}

	return eg.Wait()
}

// update 100 records
func (repo *Repository) updateRecords(ctx context.Context, appID int, rs []*Record, updateKey string) error {
	if appID == 0 {
		return errors.New("appID is required")
	}

	if len(rs) == 0 {
		return nil
	}

	select {
	case repo.Token <- struct{}{}: // acquire token
		defer func() {
			<-repo.Token
		}()
	case <-ctx.Done(): // cancelled
		return errors.New("canceled")
	}

	type UpdateRecord interface{}

	type UpdateRecordWithID struct {
		ID     string `json:"id"`
		Record Fields `json:"record"`
	}

	type UpdateKey struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}

	type UpdateRecordWithUpdateKey struct {
		UpdateKey UpdateKey `json:"updateKey"`
		Record    Fields    `json:"record"`
	}

	type requestBody struct {
		App     int            `json:"app"`
		Records []UpdateRecord `json:"records"`
	}

	records := make([]UpdateRecord, len(rs))

	for i, r := range rs {
		if updateKey == "" {
			records[i] = &UpdateRecordWithID{r.ID, r.Fields}
		} else {
			u := UpdateKey{Field: updateKey, Value: fmt.Sprint(r.Fields[updateKey])}
			delete(r.Fields, updateKey)
			records[i] = &UpdateRecordWithUpdateKey{u, r.Fields}
		}
	}

	body, err := json.Marshal(requestBody{appID, records})
	if err != nil {
		return err
	}

	_, err = repo.Client.put(APIEndpointRecords, body)
	if err != nil {
		return err
	}

	return nil
}

// update 100 records with retry
func (repo *Repository) updateRecordsWithRetry(ctx context.Context, appID int, rs []*Record, updateKey string) error {
	if appID == 0 {
		return errors.New("appID is required")
	}

	if len(rs) == 0 {
		return nil
	}

	select {
	case repo.Token <- struct{}{}: // acquire token
		defer func() {
			<-repo.Token
		}()
	case <-ctx.Done(): // cancelled
		return errors.New("canceled")
	}

	type UpdateRecord interface{}

	type UpdateRecordWithID struct {
		ID     string `json:"id"`
		Record Fields `json:"record"`
	}

	type UpdateKey struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}

	type UpdateRecordWithUpdateKey struct {
		UpdateKey UpdateKey `json:"updateKey"`
		Record    Fields    `json:"record"`
	}

	type requestBody struct {
		App     int            `json:"app"`
		Records []UpdateRecord `json:"records"`
	}

	records := make([]UpdateRecord, len(rs))

	for i, r := range rs {
		if updateKey == "" {
			records[i] = &UpdateRecordWithID{r.ID, r.Fields}
		} else {
			u := UpdateKey{Field: updateKey, Value: fmt.Sprint(r.Fields[updateKey])}
			delete(r.Fields, updateKey)
			records[i] = &UpdateRecordWithUpdateKey{u, r.Fields}
		}
	}

	body, err := json.Marshal(requestBody{appID, records})
	if err != nil {
		return err
	}

	var retryCount int

	for {
		_, err = repo.Client.put(APIEndpointRecords, body)
		if err == nil {
			break
		}

		retryCount++
		if retryCount > repo.MaxRetry {
			break
		}
		log.Printf("retry %d", retryCount)
		time.Sleep(time.Second * RetryInterval)
	}

	return err
}

//-UpdateRecords

//+DeleteRecords

func (repo *Repository) DeleteRecords(ctx context.Context, appID int, ids []string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	sliceIDs := func(ids []string, i int) [][]string {
		out := make([][]string, 0)
		for {
			if len(ids) > i {
				out = append(out, ids[:i])
			} else {
				out = append(out, ids)
			}

			if len(ids) > i {
				ids = ids[i:]
			} else {
				break
			}
		}
		return out
	}

	sliced := sliceIDs(ids, 100)

	eg, ctx := errgroup.WithContext(ctx)
	for _, _ids := range sliced {
		_ids := _ids
		eg.Go(func() error {
			return func() error {
				err := repo.deleteRecords(ctx, appID, _ids)
				if err != nil {
					return err
				}
				return nil
			}()
		})
	}

	return eg.Wait()
}

// delete 100 records
func (repo *Repository) deleteRecords(ctx context.Context, appID int, ids []string) error {
	select {
	case repo.Token <- struct{}{}: // acquire token
		defer func() {
			<-repo.Token
		}()
	case <-ctx.Done(): // cancelled
		return errors.New("canceled")
	}

	type requestBody struct {
		App int      `json:"app"`
		IDs []string `json:"ids"`
	}

	body, err := json.Marshal(requestBody{appID, ids})
	if err != nil {
		return err
	}

	_, err = repo.Client.delete(APIEndpointRecords, body)
	if err != nil {
		return err
	}

	return nil
}

//-DeleteRecords

//+UpsertRecord
func (repo *Repository) UpsertRecord(ctx context.Context, appID int, updateKey string, r *Record) (string, error) {
	if appID == 0 {
		return "", errors.New("appID is required")
	}

	if r == nil {
		return "", nil
	}

	//+既存レコードが存在するか確認
	var keyName string
	var keyValue string

	switch updateKey {
	case "":
		keyName = "レコード番号"
		keyValue = r.ID
	default:
		keyName = updateKey
		if r.Fields[updateKey] == nil {
			return "", errors.New("キーがレコードに存在しません")
		}
		keyValue = fmt.Sprint(r.Fields[updateKey])
	}

	condition := fmt.Sprintf(`%s="%s"`, keyName, keyValue)

	q := &Query{AppID: appID, Fields: []string{keyName}, Condition: condition}
	_rs, err := repo.ReadRecordsWithCursor(q)
	if err != nil {
		return "", errors.Wrap(err, "read exist key values failed")
	}
	//-既存レコードが存在するか確認

	switch len(_rs) {
	case 0:
		return repo.AddRecord(ctx, appID, r)
	case 1:
		err = repo.UpdateRecord(ctx, appID, updateKey, r)
		return r.ID, err
	default:
		return "", errors.New("Upsertのキーが複数存在します")
	}
}

//-UpsertRecord

//+UpsertRecords
func (repo *Repository) UpsertRecords(ctx context.Context, appID int, updateKey string, rs ...*Record) error {
	if ctx == nil {
		ctx = context.Background()
	}

	sliced := sliceRecords(rs, 100)

	eg, ctx := errgroup.WithContext(ctx)

	for _, _rs := range sliced {
		_rs := _rs
		eg.Go(func() error {
			return func() error {
				err := repo.upsertRecords(ctx, appID, updateKey, _rs...)
				if err != nil {
					return err
				}
				return nil
			}()
		})
	}

	return eg.Wait()
}

// upsert 100 records
func (repo *Repository) upsertRecords(ctx context.Context, appID int, updateKey string, rs ...*Record) error {
	if appID == 0 {
		return errors.New("appID is required")
	}

	if len(rs) == 0 {
		return nil
	}

	//+existKeys
	keyName := "レコード番号"
	if updateKey != "" {
		keyName = updateKey
	}

	condition := ""

	for i, r := range rs {
		keyValue := r.ID
		if updateKey != "" {
			keyValue = fmt.Sprint(r.Fields[updateKey])
		}

		if i == 0 {
			condition = fmt.Sprintf(`%s="%s"`, keyName, keyValue)
			continue
		}

		condition += fmt.Sprintf(` or %s="%s"`, keyName, keyValue)
	}

	q := &Query{AppID: appID, Fields: []string{keyName}, Condition: condition}
	_rs, err := repo.ReadRecordsWithCursor(q)
	if err != nil {
		return errors.Wrap(err, "read exist key values failed")
	}

	existKeys := make([]string, len(_rs))
	for i, r := range _rs {
		key := r.ID
		if updateKey != "" {
			key = fmt.Sprint(r.Fields[updateKey])
		}
		existKeys[i] = key
	}
	//-existKeys

	//+新規レコードと既存レコードに分類
	var addRecords []*Record
	var updateRecords []*Record

	// 既存のIDかどうかの判定
	isExistKey := func(keyValue string) bool {
		for _, k := range existKeys {
			if keyValue == k {
				return true
			}
		}
		return false
	}

	for _, r := range rs {
		keyValue := r.ID
		if updateKey != "" {
			keyValue = fmt.Sprint(r.Fields[updateKey])
		}
		if isExistKey(keyValue) {
			updateRecords = append(updateRecords, r)
			continue
		}
		addRecords = append(addRecords, r)
	}
	//-新規レコードと既存レコードに分類

	if addRecords != nil {
		_, err := repo.AddRecords(ctx, appID, addRecords...)
		if err != nil {
			return errors.Wrap(err, "add records failed")
		}
	}

	if updateRecords != nil {
		err := repo.UpdateRecords(ctx, appID, updateKey, updateRecords...)
		if err != nil {
			return errors.Wrap(err, "update records failed")
		}
	}

	return nil
}

//-UpsertRecords

// ReadFormFields ...
func (repo *Repository) ReadFormFields(appID int) (FormFields, error) {
	data, err := repo.Client.get(APIEndpointFormField, &Query{AppID: appID})
	if err != nil {
		return nil, err
	}
	raw := struct {
		Properties FormFields `json:"properties"`
	}{}

	var buf bytes.Buffer
	err = json.Indent(&buf, data, "", "  ")
	if err != nil {
		panic(err)
	}
	log.Println(buf.String())

	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	return raw.Properties, nil
}

// ReadFormLayout ...
func (repo *Repository) ReadFormLayout(appID int) (FormLayouts, error) {
	data, err := repo.Client.get(APIEndpointFormLayout, &Query{AppID: appID})
	if err != nil {
		return nil, err
	}
	raw := struct {
		Revision string        `json:"revision"`
		Layout   []*FormLayout `json:"layout"`
	}{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	return raw.Layout, nil
}

func (repo *Repository) ReadSpace(spaceID int) (*Space, error) {
	data, err := repo.Client.get(APIEndpointSpace, &Query{ID: spaceID})
	if err != nil {
		return nil, err
	}

	var s Space
	err = json.Unmarshal(data, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (repo *Repository) AddSpace(s *CreateSpace) (int, error) {
	body, err := json.Marshal(s)
	if err != nil {
		return 0, err
	}

	data, err := repo.Client.post(APIEndpointCreateSpace, body)
	if err != nil {
		return 0, err
	}

	res := struct {
		ID int `json:"id,string"`
	}{}

	err = json.Unmarshal(data, &res)
	if err != nil {
		return 0, err
	}

	return res.ID, nil
}

//+query

// Query ...
type Query struct {
	ID    int
	AppID int

	Condition string
	OrderBy   string
	limit     int
	offset    int

	Fields     []string
	TotalCount bool
}

func NewQuery(appID int) *Query {
	return &Query{AppID: appID}
}

func (q Query) String() string {
	values := url.Values{}

	if q.AppID != 0 {
		values.Add("app", strconv.Itoa(q.AppID))
		// str := fmt.Sprintf("app=%d", q.AppID)
	}

	//+id
	if q.ID != 0 {
		values.Add("id", strconv.Itoa(q.ID))
	}
	//-id

	//+query
	query := q.Condition

	if q.OrderBy != "" {
		query = fmt.Sprintf("%s order by %s", query, q.OrderBy)
	}
	if q.limit != 0 {
		query = fmt.Sprintf("%s limit %d", query, q.limit)
	}
	if q.offset != 0 {
		query = fmt.Sprintf("%s offset %d", query, q.offset)
	}

	if query != "" {
		// str = fmt.Sprintf("%s&query=%s", str, query)
		values.Add("query", query)
	}
	//-query

	if q.TotalCount {
		// str = fmt.Sprintf("%s&totalCount=true", str)
		values.Add("totalCount", "true")
	}
	if len(q.Fields) > 0 {
		for _, f := range q.Fields {
			// str = fmt.Sprintf(`%s&fields=%s`, str, f)
			values.Add("fields", f)
		}
	}
	return values.Encode()
}

//-query

//+kintone error

type resError struct {
	Code    string      `json:"code"`
	ID      string      `json:"id"`
	Message string      `json:"message"`
	Details interface{} `json:"errors"`
}

func (e *resError) Error() string {
	return fmt.Sprintf("code: %s, id: %s, message: %s, details: %s", e.Code, e.ID, e.Message, e.Details)
}

//-kintone error
