package kintone

// Webhook ...
type Webhook struct {
	Type   string `json:"type"` // ADD_RECORD or UPDAATE_RECORD
	Record Record `json:"record"`
	App    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"app"`
}
