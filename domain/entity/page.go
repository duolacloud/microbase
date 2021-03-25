package entity

type PageQuery struct {
	Filter   map[string]interface{} `json:"filter"`
	PageNo   int                    `json:"pageNo"`
	PageSize int                    `json:"pageSize"`
	Orders   []*Order               `json:"order"`
}
