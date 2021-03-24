package entity

import (
	"encoding/json"

	"github.com/duolacloud/microbase/proto/pagination"
	"github.com/thoas/go-funk"
	"google.golang.org/protobuf/types/known/wrapperspb"
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

func ConnectionQueryFromPB(q *pagination.ConnectionQuery) *ConnectionQuery {
	var filter map[string]interface{}
	if len(q.Filter) != 0 {
		_ = json.Unmarshal([]byte(q.Filter), &filter)
	}

	var first *int
	var _first int
	if q.First != nil {
		_first = int(q.First.GetValue())
		first = &_first
	}

	var last *int
	var _last int
	if q.Last != nil {
		_last = int(q.Last.GetValue())
		last = &_last
	}

	var after *string
	if q.After != nil {
		_after := q.After.GetValue()
		after = &_after
	}

	var before *string
	if q.Before != nil {
		_before := q.Before.GetValue()
		before = &_before
	}

	return &ConnectionQuery{
		First:     first,
		Last:      last,
		Before:    before,
		After:     after,
		Filter:    filter,
		Fields:    q.Fields,
		NeedTotal: q.NeedTotal,
		Orders: funk.Map(q.Orders, func(o *pagination.Order) *Order {
			var direction OrderDirection
			if o.Direction == pagination.OrderDirection_DESC {
				direction = OrderDirectionDesc
			} else {
				direction = OrderDirectionAsc
			}

			return &Order{
				Field:     o.Field,
				Direction: direction,
			}
		}).([]*Order),
	}
}

func (c *ConnectionQuery) ToPB() *pagination.ConnectionQuery {
	filterB, _ := json.Marshal(c.Filter)

	pbquery := &pagination.ConnectionQuery{
		NeedTotal: c.NeedTotal,
		Fields:    c.Fields,
		Filter:    string(filterB),
		Orders: funk.Map(c.Orders, func(o *Order) *pagination.Order {
			var direction pagination.OrderDirection
			if o.Direction == OrderDirectionDesc {
				direction = pagination.OrderDirection_DESC
			} else {
				direction = pagination.OrderDirection_ASC
			}

			return &pagination.Order{
				Field:     o.Field,
				Direction: direction,
			}
		}).([]*pagination.Order),
	}

	if c.First != nil {
		pbquery.First = wrapperspb.Int32(int32(*c.First))
	}

	if c.Last != nil {
		pbquery.Last = wrapperspb.Int32(int32(*c.Last))
	}

	if c.Before != nil {
		pbquery.Before = wrapperspb.String(*c.Before)
	}

	if c.After != nil {
		pbquery.After = wrapperspb.String(*c.After)
	}

	return pbquery
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

func (p *PageInfo) ToPB() *pagination.PageInfo {
	return &pagination.PageInfo{
		HasPrevious: p.HasPrevious,
		HasNext:     p.HasNext,
		EndCursor:   p.EndCursor,
		StartCursor: p.StartCursor,
	}
}
