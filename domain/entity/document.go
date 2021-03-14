package entity

type Document struct {
	Index  string                 `json:"index"`
	Type   string                 `json:"type"`
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}
