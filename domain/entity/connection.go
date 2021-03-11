package entity

import (
	"encoding/json"

	"github.com/duolacloud/microbase/proto/api"
	"github.com/thoas/go-funk"
)

type ConnectionQuery struct {
	Filter    map[string]interface{} `json:"filter"` // 筛选条件
	First     *int                   `json:"first"`
	Last      *int                   `json:"last"`
	Before    *string                `json:"before"`
	After     *string                `json:"after"`
	Fields    []string               `json:"fields"`
	Orders    []*Order               `json:"order"` // 游标字段&排序
	NeedTotal bool                   `json:"needTotal"`
}

func (c *ConnectionQuery) FromPB(q *api.ConnectionQuery) {
	var filter map[string]interface{}
	if len(q.Filter) != 0 {
		_ = json.Unmarshal([]byte(q.Filter), &filter)
	}

	var first *int
	var _first int
	if q.HasFirst {
		_first = int(q.First)
		first = &_first
	}

	var last *int
	var _last int
	if q.HasLast {
		_last = int(q.Last)
		last = &_last
	}

	var after *string
	if q.HasAfter {
		after = &q.After
	}

	var before *string
	if q.HasBefore {
		before = &q.Before
	}

	c.First = first
	c.Last = last
	c.Before = before
	c.After = after
	c.Filter = filter
	c.Fields = q.Fields
	c.NeedTotal = q.NeedTotal
	c.Orders = funk.Map(q.Orders, func(o *api.OrderSpec) *Order {
		var direction OrderDirection
		if o.Direction == api.OrderDirection_DESC {
			direction = OrderDirectionDesc
		} else {
			direction = OrderDirectionAsc
		}

		return &Order{
			Field:     o.Field,
			Direction: direction,
		}
	}).([]*Order)
}

type Edge struct {
	Node   interface{} `json:"node"`
	Cursor string      `json:"cursor"`
}

type Connection struct {
	Total    int64    `json:"total"`
	Edges    []*Edge  `json:"edges"`
	PageInfo PageInfo `json:"pageInfo"`
}

type PageInfo struct {
	HasPrevious bool   `json:"hasPrevious"` // 是否有更多数据
	HasNext     bool   `json:"hasNext"`     // 是否有更多数据
	EndCursor   string `json:"endCursor"`   // 结果集中的起始游标值
	StartCursor string `json:"startCursor"` // 结果集中的结束游标值
}

func (p *PageInfo) ToPB() *api.PageInfo {
	return &api.PageInfo{
		HasPrevious: p.HasPrevious,
		HasNext:     p.HasNext,
		EndCursor:   p.EndCursor,
		StartCursor: p.StartCursor,
	}
}
