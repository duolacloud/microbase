package elasticsearch

import (
	"context"
	"errors"
	"fmt"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"
	"github.com/duolacloud/microbase/utils/xtree"
	"github.com/olivere/elastic/v6"
	"github.com/thoas/go-funk"
)

type queryVisitor struct {
	key   string
	value interface{}
	query elastic.Query
}

func (n *queryVisitor) children(current *xtree.Node) error {
	var hasChildren bool

	if current.Parent == nil {
		hasChildren = true
	} else {
		filterType := entity.FilterType(n.key)

		switch filterType {
		case entity.FilterType_AND, entity.FilterType_OR:
			hasChildren = true
		}
	}

	if !hasChildren {
		return nil
	}

	vMap, ok := n.value.(map[string]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("child query key format not match"))
	}

	children := make([]*xtree.Node, 0, len(vMap))
	for key, child := range vMap {
		child := &xtree.Node{
			Visitor: &queryVisitor{
				key:   key,
				value: child,
			},
			Parent: current,
		}

		children = append(children, child)
	}
	current.Children = children
	return nil
}

func (n *queryVisitor) PreVisit(c context.Context, current *xtree.Node) error {
	err := n.children(current)
	if err != nil {
		return err
	}

	if current.Parent == nil {
		n.query = elastic.NewBoolQuery()
		return nil
	}

	filterType := entity.FilterType(n.key)

	switch filterType {
	case entity.FilterType_AND:
		{
			n.query = elastic.NewBoolQuery()
			return nil
		}
	case entity.FilterType_OR:
		{
			n.query = elastic.NewBoolQuery()
			return nil
		}
	}

	fieldName := n.key

	vMap, ok := n.value.(map[string]interface{})
	if !ok {
		n.query = elastic.NewTermQuery(fieldName, n.value)
		return nil
	}

	rangeQuery := elastic.NewRangeQuery(fieldName)
	isRangeQuery := false

	for vKey, vValue := range vMap {
		/*
			switch field.Struct.Type.String() {
			case "time.Time", "*time.Time":
				v, err := smarttime.Parse(vValue)
				if err == nil {
					vValue = v
				}
			}
		*/

		filterType = entity.FilterType(vKey)
		switch filterType {
		case entity.FilterType_EQ:
			elastic.NewTermQuery(fieldName, vValue)
		case entity.FilterType_NE:
			n.query = elastic.NewTermQuery(fieldName, vValue)
		case entity.FilterType_GT:
			rangeQuery.Gt(vValue)
			isRangeQuery = true
		case entity.FilterType_GTE:
			rangeQuery.Gte(vValue)
			isRangeQuery = true
		case entity.FilterType_LT:
			rangeQuery.Lt(vValue)
			isRangeQuery = true
		case entity.FilterType_LTE:
			rangeQuery.Lte(vValue)
			isRangeQuery = true
		case entity.FilterType_LIKE:
			n.query = elastic.NewTermQuery(fieldName, vValue)
		case entity.FilterType_MATCH:
			n.query = elastic.NewTermQuery(fieldName, vValue)
		case entity.FilterType_NOT_LIKE:
			n.query = elastic.NewTermQuery(fieldName, vValue)
		case entity.FilterType_IN:
			n.query = elastic.NewTermQuery(fieldName, vValue)
		case entity.FilterType_NOT_IN:
			n.query = elastic.NewTermQuery(fieldName, vValue)
		case entity.FilterType_BETWEEN:
			values, ok := n.value.([]interface{})
			if !ok {
				return repository.ErrFilterValueType
			}

			if len(values) != 2 {
				return repository.ErrFilterValueSize
			}
			rangeQuery.Lte(values[0]).Gte(values[1])
			isRangeQuery = true
		case entity.FilterType_IS_NULL:
			n.query = elastic.NewTermQuery(fieldName, vValue)
		case entity.FilterType_NOT_NULL:
			n.query = elastic.NewTermQuery(fieldName, vValue)
		}
	}

	if isRangeQuery {
		n.query = rangeQuery
	}

	return nil
}

func (n *queryVisitor) PostVisit(c context.Context, current *xtree.Node) error {
	if current.Parent == nil {
		subQueries := funk.Map(current.Children, func(child *xtree.Node) elastic.Query {
			subVisitor := child.Visitor.(*queryVisitor)
			return subVisitor.query
		}).([]elastic.Query)
		boolQuery := n.query.(*elastic.BoolQuery)
		boolQuery.Filter(subQueries...)

		return nil
	}

	filterType := entity.FilterType(n.key)
	for _, child := range current.Children {
		childVisitor := child.Visitor.(*queryVisitor)
		query := n.query.(*elastic.BoolQuery)

		switch filterType {
		case entity.FilterType_EQ,
			entity.FilterType_GT,
			entity.FilterType_GTE,
			entity.FilterType_LT,
			entity.FilterType_LTE,
			entity.FilterType_LIKE,
			entity.FilterType_MATCH,
			entity.FilterType_IN,
			entity.FilterType_BETWEEN,
			entity.FilterType_IS_NULL:
			query.Must(childVisitor.query)
		case entity.FilterType_NE, entity.FilterType_NOT_LIKE, entity.FilterType_NOT_IN, entity.FilterType_NOT_NULL:
			query.MustNot(childVisitor.query)
		}
	}
	return nil
}

func applyFilter(c context.Context, filter map[string]interface{}) (elastic.Query, error) {
	if filter == nil || len(filter) == 0 {
		return elastic.NewBoolQuery(), nil
	}

	rootVisitor := &queryVisitor{
		value: filter,
	}

	tree := xtree.NewXTree()
	if err := tree.Travel(c, &xtree.Node{
		Visitor: rootVisitor,
	}); err != nil {
		return nil, err
	}

	return rootVisitor.query, nil
}
