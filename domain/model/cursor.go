package model

type CursorQuery struct {
	Filter     map[string]interface{} `json:"filter"`     // 筛选条件
	Cursor     interface{}            `json:"cursor"`     // 游标值
	CursorSort *SortSpec              `json:"cursorSort"` // 游标字段&排序
	Size       int                    `json:"size"`       // 数据量
	Direction  byte                   `json:"direction"`  // 查询方向 0：游标前；1：游标后
}

type CursorExtra struct {
	Direction   byte        `json:"direction"`   // 查询方向 0：游标前；1：游标后
	Total       int         `json:"total"`       // 数据量
	HasPrevious bool        `json:"hasPrevious"` // 是否有更多数据
	HasNext     bool        `json:"hasNext"`     // 是否有更多数据
	EndCursor   interface{} `json:"maxCursor"`   // 结果集中的起始游标值
	StartCursor interface{} `json:"minCursor"`   // 结果集中的结束游标值
}
