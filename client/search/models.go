package search

type Index struct {
	Name    string                 `json:"name"`
	Mapping map[string]interface{} `json:"mappings"`
}

type Document struct {
	Index  string                 `json:"index"`
	Type   string                 `json:"type"`
	Fields map[string]interface{} `json:"fields"`
}
