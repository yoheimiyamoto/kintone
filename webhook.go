package kintone

import "regexp"

// Webhook ...
type Webhook struct {
	Type   string `json:"type"` // ADD_RECORD or UPDAATE_RECORD
	Record Record `json:"record"`
	App    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"app"`
	RecordTitle string
	URL         URL
}

type URL string

func (u URL) Subdomain() string {
	r := regexp.MustCompile(`^https://(.+)\.cybozu\.com`)
	result := r.FindStringSubmatch(string(u))
	if len(result) < 2 {
		return ""
	}
	return result[1]
}
