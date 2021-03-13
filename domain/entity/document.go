package entity

type Document struct {
	Index string                 `json:"index"`
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Data  map[string]interface{} `json:"data"`
}
