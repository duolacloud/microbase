package entity

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/duolacloud/microbase/proto/pagination"
	"github.com/thoas/go-funk"
	"github.com/vmihailenco/msgpack/v5"
)

type Cursor struct {
	// ID    interface{} `msgpack:"i"`
	Value []interface{} `msgpack:"v"`
}

func (c *Cursor) Unmarshal(s string) error {
	if err := msgpack.NewDecoder(
		base64.NewDecoder(
			base64.RawStdEncoding,
			strings.NewReader(s),
		),
	).Decode(c); err != nil {
		return fmt.Errorf("cannot decode cursor: %w", err)
	}
	return nil
}

func (c *Cursor) Marshal(w io.Writer) error {
	wc := base64.NewEncoder(base64.RawStdEncoding, w)
	defer wc.Close()
	return msgpack.NewEncoder(wc).Encode(c)
}

type CursorDirection string

const (
	CursorDirectionAfter  CursorDirection = "AFTER"
	CursorDirectionBefore CursorDirection = "BEFORE"
)

type CursorQuery struct {
	Filter    map[string]interface{} `json:"filter"`    // 筛选条件
	Cursor    string                 `json:"cursor"`    // 游标值
	Orders    []*Order               `json:"order"`     // 游标字段&排序
	Size      int                    `json:"size"`      // 数据量
	Direction CursorDirection        `json:"direction"` // 查询方向 0：游标前；1：游标后
	NeedTotal bool                   `json:"needTotal"`
	Fields    []string               `json:"fields"`
}

func (q *CursorQuery) FromPB(o *pagination.ListQuery) {
	var filter map[string]interface{}
	_ = json.Unmarshal([]byte(o.Filter), &filter)

	if o.Direction == pagination.CursorDirection_before {
		q.Direction = CursorDirectionBefore
	} else {
		q.Direction = CursorDirectionAfter
	}

	q.Filter = filter
	q.Cursor = o.Cursor

	if o.Orders != nil {
		q.Orders = funk.Map(o.Orders, func(o pagination.Order) *Order {
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
		}).([]*Order)
	}

	q.Size = int(o.Size)
	q.NeedTotal = o.NeedTotal
	q.Fields = o.Fields
}

type CursorExtra struct {
	Total       int64  `json:"total"`       // 数据量
	HasPrevious bool   `json:"hasPrevious"` // 是否有更多数据
	HasNext     bool   `json:"hasNext"`     // 是否有更多数据
	EndCursor   string `json:"endCursor"`   // 结果集中的起始游标值
	StartCursor string `json:"startCursor"` // 结果集中的结束游标值
}
