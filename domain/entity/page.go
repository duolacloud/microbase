package entity

type PageQuery struct {
	Filter   map[string]interface{} `json:"filter"`
	PageNo   int                    `json:"pageNo"`
	PageSize int                    `json:"pageSize"`
	Orders   []*Order               `json:"order"`
}

type Page struct {
	Content   interface{} `json:"content"`
	Total     int         `json:"total"`
	PageNo    int         `json:"pageNo"`
	PageSize  int         `json:"pageSize"`
	PageCount int         `json:"pageCount"`
}
