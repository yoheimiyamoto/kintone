package kintone

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/sync/errgroup"
)

// Repository ...
type Repository struct {
	Client Client
	Token  chan struct{}
}

type RepositoryOption struct {
	HTTPClient    *http.Client
	MaxConcurrent int
}

// NewRepository ...
func NewRepository(subdomain string, username, password string, option *RepositoryOption) *Repository {
	var httpClient *http.Client
	maxConcurrent := 1

	if option != nil {
		httpClient = option.HTTPClient
		maxConcurrent = option.MaxConcurrent
	}

	c := newClient(subdomain, username, password, httpClient)
	token := make(chan struct{}, maxConcurrent)
	u, _ := url.ParseRequestURI(fmt.Sprintf(APIEndpointBase, subdomain))
	c.endpointBase = u
	return &Repository{c, token}
}

// ReadRecords ...
func (repo *Repository) ReadRecords(ctx context.Context, q *Query) ([]*Record, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// レコード数確認
	totalCount, err := repo.readTotalCount(q)
	if err != nil {
		return nil, err
	}

	recordsCh := make(chan *Record, totalCount)

	q.RawQuery.Limit = 500

	eg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < totalCount; i += 500 {

		// クエリ生成（コピー）
		q := *q
		raw := *q.RawQuery
		raw.Offset = i
		q.RawQuery = &raw
		eg.Go(func() error {
			_rs, err := repo.read500Records(ctx, &q)
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

func (repo *Repository) read500Records(ctx context.Context, q *Query) ([]*Record, error) {
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
		// Records    []*Record `json:"records"`
		TotalCount int `json:"totalCount,string"`
	}{}
	if err := json.Unmarshal(body, &r); err != nil {
		return 0, err
	}

	return r.TotalCount, nil
}

//+AddRecord

// AddRecords ...
func (repo *Repository) AddRecords(ctx context.Context, appID string, rs []*Record) ([]string, error) {
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
				_ids, err := repo.addRecords(ctx, appID, _rs)
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

func (repo *Repository) addRecords(ctx context.Context, appID string, rs []*Record) ([]string, error) {
	select {
	case repo.Token <- struct{}{}: // acquire token
		defer func() {
			<-repo.Token
		}()
	case <-ctx.Done(): // cancelled
		return nil, errors.New("canceled")
	}

	type requestBody struct {
		App     string   `json:"app"`
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

//-AddRecord

//+UpdateRecords

// UpdateRecords ...
func (repo *Repository) UpdateRecords(ctx context.Context, appID string, updateKey string, rs ...*Record) error {
	if ctx == nil {
		ctx = context.Background()
	}

	sliced := sliceRecords(rs, 100)

	eg, ctx := errgroup.WithContext(ctx)
	for _, _rs := range sliced {
		_rs := _rs
		eg.Go(func() error {
			return func() error {
				err := repo.updateRecords(ctx, appID, _rs, updateKey)
				if err != nil {
					return err
				}
				return nil
			}()
		})
	}

	return eg.Wait()
}

// 100レコードづつUpdate
func (repo *Repository) updateRecords(ctx context.Context, appID string, rs []*Record, updateKeyCode string) error {
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
		App     string         `json:"app"`
		Records []UpdateRecord `json:"records"`
	}

	records := make([]UpdateRecord, len(rs))

	for i, r := range rs {
		if updateKeyCode == "" {
			records[i] = &UpdateRecordWithID{r.ID, r.Fields}
		} else {
			updateKey := UpdateKey{Field: updateKeyCode, Value: fmt.Sprint(r.Fields[updateKeyCode])}
			delete(r.Fields, updateKeyCode)
			records[i] = &UpdateRecordWithUpdateKey{updateKey, r.Fields}
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

//-UpdateRecords

//+DeleteRecords

// DeleteRecords ...
func (repo *Repository) DeleteRecords(ctx context.Context, appID int, ids []string) error {
	if ctx == nil {
		ctx = context.Background()
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

// 100レコードづつDelete
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

func sliceIDs(ids []string, i int) [][]string {
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

//-DeleteRecords

// ReadFormFields ...
func (repo *Repository) ReadFormFields(appID string) (FormFields, error) {
	data, err := repo.Client.get(APIEndpointFormField, &Query{AppID: appID})
	if err != nil {
		return nil, err
	}
	raw := struct {
		Properties FormFields `json:"properties"`
	}{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	return raw.Properties, nil
}

// ReadFormLayout ...
func (repo *Repository) ReadFormLayout(appID string) (FormLayouts, error) {
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

//+query

// type Query string

// Query ...
type Query struct {
	AppID      string
	RawQuery   *RawQuery
	Fields     []string
	TotalCount bool
}

// RawQuery ...
type RawQuery struct {
	Condition string
	OrderBy   string
	Limit     int
	Offset    int
}

func (r RawQuery) String() string {
	str := r.Condition
	if r.OrderBy != "" {
		str = fmt.Sprintf("%s offset %s", str, r.OrderBy)
	}
	if r.Limit != 0 {
		str = fmt.Sprintf("%s limit %d", str, r.Limit)
	}
	if r.Offset != 0 {
		str = fmt.Sprintf("%s offset %d", str, r.Offset)
	}
	return strings.TrimSpace(str)
}

func (q Query) String() string {
	str := fmt.Sprintf("app=%s", q.AppID)
	if q.RawQuery != nil {
		str = fmt.Sprintf("%s&query=%s", str, q.RawQuery)
	}
	if q.TotalCount {
		str = fmt.Sprintf("%s&totalCount=true", str)
	}
	if len(q.Fields) > 0 {
		for _, f := range q.Fields {
			str = fmt.Sprintf(`%s&fields=%s`, str, f)
		}
	}
	return str
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
