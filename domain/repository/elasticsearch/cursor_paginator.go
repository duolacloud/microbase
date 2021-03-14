package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/olivere/elastic/v6"
	// "github.com/olivere/elastic/v7"
)

type CursorPaginator struct {
	client *elastic.Client
}

func NewCursorPaginator(client *elastic.Client) *CursorPaginator {
	return &CursorPaginator{
		client,
	}
}

func (p *CursorPaginator) Paginate(c context.Context, query *entity.CursorQuery, index, typ string) ([]*entity.Document, *entity.CursorExtra, error) {
	filter, err := applyFilter(c, query.Filter)
	if err != nil {
		return nil, nil, err
	}

	extra := &entity.CursorExtra{}

	if query.NeedTotal {
		total, err := p.client.Count().
			Index(index).
			Type(typ).
			Query(filter).
			Do(c)
		if err != nil {
			return nil, nil, err
		}
		extra.Total = int(total)
	}

	limit := query.Size + 1

	p.ensureOrders(query)

	searchService := p.client.Search().
		Index(index).
		Type(typ).
		Size(limit)

	cursorFilters, err := p.applyCursor(query)
	if err != nil {
		return nil, nil, err
	}

	rootFilters := make([]elastic.Query, 0)
	rootFilters = append(rootFilters, filter)
	rootFilters = append(rootFilters, cursorFilters...)

	searchService.Query(elastic.NewBoolQuery().Filter(rootFilters...))
	p.applyOrders(searchService, query.Orders, query.Direction == entity.CursorDirectionBefore)

	result, err := searchService.Do(c)
	if err != nil {
		return nil, nil, err
	}

	if len(result.Hits.Hits) == 0 {
		return nil, &entity.CursorExtra{
			HasNext:     false,
			HasPrevious: false,
			Total:       0,
		}, nil
	}

	docs := make([]*entity.Document, len(result.Hits.Hits))
	for i, r := range result.Hits.Hits {
		doc := &entity.Document{
			Index: index,
			Type:  typ,
		}

		err = json.Unmarshal(*r.Source, &doc.Fields)
		if err != nil {
			return nil, nil, err
		}

		docs[i] = doc
	}

	if limit == len(docs) {
		extra.HasNext = true
		extra.HasPrevious = true
		docs = docs[:limit-1]
	}

	toCursor := func(doc *entity.Document) (string, error) {
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

	extra.StartCursor, err = toCursor(docs[0])
	if err != nil {
		return nil, nil, err
	}

	extra.EndCursor, err = toCursor(docs[len(docs)-1])
	if err != nil {
		return nil, nil, err
	}

	return docs, extra, nil
}

func (p *CursorPaginator) applyCursor(query *entity.CursorQuery) ([]elastic.Query, error) {
	if len(query.Cursor) > 0 {
		cursor := &entity.Cursor{}
		err := cursor.Unmarshal(query.Cursor)
		if err != nil {
			return nil, err
		}

		if len(cursor.Value) == 0 {
			return nil, nil
		}

		if len(cursor.Value) != len(query.Orders) {
			return nil, errors.New(fmt.Sprintf("cursor format fields length: %d not match orders fields length: %d", len(cursor.Value), len(query.Orders)))
		}

		filters := make([]elastic.Query, len(cursor.Value))

		for i, value := range cursor.Value {
			order := query.Orders[i]

			if query.Direction == entity.CursorDirectionAfter {
				// after
				if query.Orders[0].Direction == entity.OrderDirectionDesc {
					filters[i] = elastic.NewRangeQuery(order.Field).Lt(value)
				} else {
					filters[i] = elastic.NewRangeQuery(order.Field).Gt(value)
				}
			} else {
				// before
				if query.Orders[0].Direction == entity.OrderDirectionDesc {
					filters[i] = elastic.NewRangeQuery(order.Field).Gt(value)
				} else {
					filters[i] = elastic.NewRangeQuery(order.Field).Lt(value)
				}
			}
		}

		return filters, nil
	}

	return nil, nil
}

func (p *CursorPaginator) applyOrders(searchService *elastic.SearchService, orders []*entity.Order, reverse bool) {
	for _, order := range orders {
		if reverse {
			order.Direction = orders[0].Direction.Reverse()
		}

		asc := order.Direction != entity.OrderDirectionDesc

		searchService.Sort(order.Field, asc)
	}
}

func (p *CursorPaginator) ensureOrders(query *entity.CursorQuery) {
	// 排序一定要包含ID, 否则会遍历不完整
	matchCount := 0
	for _, order := range query.Orders {
		if order.Field == "id" {
			matchCount += 1
		}
	}

	if matchCount != 1 {
		if query.Orders == nil {
			query.Orders = make([]*entity.Order, 0)
		}

		order := &entity.Order{
			Field:     "id",
			Direction: entity.OrderDirectionAsc,
		}
		query.Orders = append(query.Orders, order)
	}
}
