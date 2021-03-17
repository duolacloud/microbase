package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/duolacloud/microbase/client/search"
	"github.com/duolacloud/microbase/domain/entity"
	"github.com/olivere/elastic/v6"
)

type ConnectionPaginator struct {
	client *elastic.Client
}

func NewConnectionPaginator(client *elastic.Client) *ConnectionPaginator {
	return &ConnectionPaginator{
		client,
	}
}

func validateFirstLast(first, last *int) error {
	switch {
	case first != nil && last != nil:
		return errors.New("Passing both `first` and `last` to paginate a connection is not supported.")
	case first != nil && *first < 0:
		return errors.New("`first` on a connection cannot be less than zero.")
	case last != nil && *last < 0:
		return errors.New("`last` on a connection cannot be less than zero.")
	}
	return nil
}

func (p *ConnectionPaginator) Paginate(c context.Context, query *entity.ConnectionQuery, index, typ string) (conn *entity.Connection, err error) {
	err = validateFirstLast(query.First, query.Last)
	if err != nil {
		return
	}

	filter, err := applyFilter(c, query.Filter)
	if err != nil {
		return nil, err
	}

	conn = &entity.Connection{}
	if query.First != nil && *query.First == 0 ||
		query.Last != nil && *query.Last == 0 {
		if query.NeedTotal {
			conn.Total, err = p.client.Count().
				Index(index).
				Type(typ).
				Query(filter).
				Do(c)

			if err != nil {
				return
			}

			conn.PageInfo.HasNext = query.First != nil && conn.Total > 0
			conn.PageInfo.HasPrevious = query.Last != nil && conn.Total > 0
		}
		return conn, nil
	}

	if query.NeedTotal {
		conn.Total, err = p.client.Count().
			Index(index).
			Type(typ).
			Query(filter).
			Do(c)
		if err != nil {
			return
		}
	}

	p.ensureOrders(&query.Orders)

	cursorFilters, err := p.applyCursor(query)
	if err != nil {
		return nil, err
	}

	var limit int
	if query.First != nil {
		limit = *query.First + 1
	} else if query.Last != nil {
		limit = *query.Last + 1
	}

	if limit == 0 {
		limit = 21
	}

	if limit > 1001 {
		limit = 1001
	}

	rootFilters := make([]elastic.Query, 0)
	rootFilters = append(rootFilters, filter)
	rootFilters = append(rootFilters, cursorFilters...)

	if len(query.Fields) > 0 {
		// TODO 只获取指定字段
	}

	searchService := p.client.Search().
		Index(index).
		Type(typ).
		Size(limit).
		Query(elastic.NewBoolQuery().Filter(rootFilters...))

	p.applyOrders(searchService, query.Orders, query.Last != nil)

	result, err1 := searchService.Do(c)
	if err1 != nil {
		err = err1
		return
	}

	if len(result.Hits.Hits) == 0 {
		return
	}

	docs := make([]*search.Document, len(result.Hits.Hits))
	for i, r := range result.Hits.Hits {
		doc := &search.Document{
			Index: index,
			Type:  typ,
		}

		err = json.Unmarshal(*r.Source, &doc.Fields)
		if err != nil {
			return
		}

		docs[i] = doc
	}

	if limit == len(docs) {
		conn.PageInfo.HasNext = true
		conn.PageInfo.HasPrevious = true
		docs = docs[:limit-1]
	}

	var nodeAt func(int) *search.Document
	if query.Last != nil {
		n := len(docs) - 1
		nodeAt = func(i int) *search.Document {
			return docs[n-i]
		}
	} else {
		nodeAt = func(i int) *search.Document {
			return docs[i]
		}
	}

	toCursor := func(doc *search.Document) (string, error) {
		orderFieldValues := make([]interface{}, len(query.Orders))
		for i, order := range query.Orders {
			v, ok := doc.Fields[order.Field]
			if !ok {
				return "", errors.New(fmt.Sprintf("cursor field %s not found within doc", order.Field))
			}

			orderFieldValues[i] = v
		}

		cursor := &entity.Cursor{
			Value: orderFieldValues,
		}

		w := new(bytes.Buffer)
		err = cursor.Marshal(w)
		if err != nil {
			return "", err
		}

		return w.String(), nil
	}

	conn.Edges = make([]*entity.Edge, len(docs))
	for i := range conn.Edges {
		node := nodeAt(i)

		var cursor string
		cursor, err = toCursor(node)
		if err != nil {
			return
		}

		conn.Edges[i] = &entity.Edge{
			Node:   node,
			Cursor: cursor,
		}
	}

	conn.PageInfo.StartCursor = conn.Edges[0].Cursor
	conn.PageInfo.EndCursor = conn.Edges[len(conn.Edges)-1].Cursor

	return
}

func (p *ConnectionPaginator) applyOrders(searchService *elastic.SearchService, orders []*entity.Order, reverse bool) {
	log.Printf("applyOrders reverse: %v", reverse)
	for _, order := range orders {
		if reverse {
			order.Direction = orders[0].Direction.Reverse()
		}

		asc := order.Direction != entity.OrderDirectionDesc

		searchService.Sort(order.Field, asc)
	}
}

func (p *ConnectionPaginator) ensureOrders(orders *[]*entity.Order) {
	// 排序一定要包含ID, 否则会遍历不完整
	matchCount := 0
	for _, order := range *orders {
		if order.Field == "id" {
			matchCount += 1
		}
	}

	if matchCount != 1 {
		if orders == nil {
			*orders = make([]*entity.Order, 0)
		}

		order := &entity.Order{
			Field:     "id",
			Direction: entity.OrderDirectionAsc,
		}
		*orders = append(*orders, order)
	}
}

func (p *ConnectionPaginator) applyCursor(query *entity.ConnectionQuery) ([]elastic.Query, error) {
	filters := make([]elastic.Query, 0)

	if query.After != nil {
		if len(*query.After) != 0 {
			cursor := &entity.Cursor{}
			err := cursor.Unmarshal(*query.After)
			if err != nil {
				return nil, err
			}

			if query.Orders != nil {
				if len(cursor.Value) != len(query.Orders) {
					return nil, errors.New(fmt.Sprintf("cursor format fields length: %d not match orders fields length: %d", len(cursor.Value), len(query.Orders)))
				}

				for i, order := range query.Orders {
					// 以第一个排序字段的顺序为准, 合并比较
					if query.Orders[0].Direction != entity.OrderDirectionDesc {
						query := elastic.NewRangeQuery(order.Field).Gt(cursor.Value[i])
						filters = append(filters, query)
					} else {
						query := elastic.NewRangeQuery(order.Field).Lt(cursor.Value[i])
						filters = append(filters, query)
					}
				}
			}
		}
	}

	if query.Before != nil {
		if len(*query.Before) != 0 {
			cursor := &entity.Cursor{}
			err := cursor.Unmarshal(*query.Before)
			if err != nil {
				return nil, err
			}

			if query.Orders != nil {
				if len(cursor.Value) != len(query.Orders) {
					return nil, errors.New(fmt.Sprintf("cursor format fields length: %d not match orders fields length: %d", len(cursor.Value), len(query.Orders)))
				}

				for i, order := range query.Orders {
					if query.Orders[0].Direction != entity.OrderDirectionDesc {
						query := elastic.NewRangeQuery(order.Field).Lt(cursor.Value[i])
						filters = append(filters, query)
					} else {
						query := elastic.NewRangeQuery(order.Field).Gt(cursor.Value[i])
						filters = append(filters, query)
					}
				}
			}
		}
	}

	return filters, nil
}
