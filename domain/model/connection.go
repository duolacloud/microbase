package model

type ConnectionQuery struct {
	Filter    map[string]interface{} `json:"filter"`   // 筛选条件
	HasFirst  bool                   `json:"hasFirst"` // 游标值
	First     int                    `json:"first"`
	HasLast   bool                   `json:"hasLast"`
	Last      string                 `json:"last"`
	HasBefore bool                   `json:"hasBefore"`
	Before    string                 `json:"before"`
	HasAfter  bool                   `json:"hasAfter"`
	After     string                 `json:"after"`

	Orders    []*Order `json:"order"` // 游标字段&排序
	NeedTotal bool     `json:"needTotal"`
}

type Edge struct {
	Node   interface{} `json:"node"`
	Cursor string      `json:"cursor"`
}

type Connection struct {
	Edges    []*Edge   `json:"edges"`
	PageInfo *PageInfo `json:"pageInfo"`
}

type PageInfo struct {
	HasPrevious bool   `json:"hasPrevious"` // 是否有更多数据
	HasNext     bool   `json:"hasNext"`     // 是否有更多数据
	EndCursor   string `json:"endCursor"`   // 结果集中的起始游标值
	StartCursor string `json:"startCursor"` // 结果集中的结束游标值
}
