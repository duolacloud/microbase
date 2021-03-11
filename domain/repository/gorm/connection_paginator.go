/**
 * Facebook 的 Graphql relay 查询模式
 */
package gorm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/duolacloud/microbase/domain/model"
	"github.com/duolacloud/microbase/domain/repository"
	"github.com/duolacloud/microbase/logger"
	breflect "github.com/duolacloud/microbase/reflect"
	"github.com/duolacloud/microbase/types/smarttime"
	_gorm "github.com/jinzhu/gorm"
)

type connectionPaginator struct {
	db          *_gorm.DB
	model       model.Model
	modelStruct *_gorm.ModelStruct
}

func NewConnectionPaginator(db *_gorm.DB, model model.Model) repository.ConnectionPaginator {
	modelStruct := db.NewScope(model).GetModelStruct()

	return &connectionPaginator{
		db,
		model,
		modelStruct,
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

func (p *connectionPaginator) Paginate(c context.Context, query *model.ConnectionQuery) (conn *model.Connection, err error) {
	if len(p.modelStruct.PrimaryFields) == 0 {
		err = errors.New("no primary key found for model")
		return
	}

	err = validateFirstLast(query.First, query.Last)
	if err != nil {
		return
	}

	if query.Orders == nil {
		fieldIDs := p.modelStruct.PrimaryFields
		query.Orders = make([]*model.Order, len(fieldIDs))
		for i, fieldID := range fieldIDs {
			order := &model.Order{
				Field:     fieldID.DBName,
				Direction: model.OrderDirectionAsc,
			}
			query.Orders[i] = order
		}
	}

	dbHandler := p.db.Model(p.model)
	dbHandler, err = applyFilter(dbHandler, p.modelStruct, query.Filter)
	if err != nil {
		return nil, err
	}

	conn = &model.Connection{Edges: []*model.Edge{}}
	if query.First != nil && *query.First == 0 ||
		query.Last != nil && *query.Last == 0 {
		if query.NeedTotal {
			if db := dbHandler.Count(&conn.Total); db.Error != nil {
				err = db.Error
				return
			}

			conn.PageInfo.HasNext = query.First != nil && conn.Total > 0
			conn.PageInfo.HasPrevious = query.Last != nil && conn.Total > 0
		}
		return conn, nil
	}

	if query.NeedTotal {
		if err = dbHandler.Count(&conn.Total).Error; err != nil {
			return
		}
	}

	dbHandler, err = p.applyCursor(dbHandler, query)
	if err != nil {
		return nil, err
	}

	dbHandler, err = p.applyOrders(dbHandler, query.Orders, query.Last != nil)
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

	dbHandler = dbHandler.Limit(limit)

	if len(query.Fields) > 0 {
		dbHandler = dbHandler.Select(query.Fields)
	}

	resultPtr := breflect.MakeSlicePtr(p.model, 0, limit)
	if err = dbHandler.Find(resultPtr).Error; err != nil {
		return
	}

	count := breflect.SlicePtrLen(resultPtr)
	if count == 0 {
		return
	}

	if count == limit {
		conn.PageInfo.HasNext = true
		conn.PageInfo.HasPrevious = true
		breflect.SlicePtrSlice3To(resultPtr, 0, limit-1, limit-1, resultPtr)
	}

	var nodeAt func(int) interface{}
	if query.Last != nil {
		n := breflect.SlicePtrLen(resultPtr) - 1
		nodeAt = func(i int) interface{} {
			return breflect.SlicePtrIndexOf(resultPtr, n-i)
		}
	} else {
		nodeAt = func(i int) interface{} {
			return breflect.SlicePtrIndexOf(resultPtr, i)
		}
	}

	toCursor := func(item interface{}) (string, error) {
		orderFieldValues := make([]interface{}, len(query.Orders))
		for i, order := range query.Orders {
			fieldOrder, ok := p.findField(order.Field, dbHandler)
			if !ok {
				return "", errors.New(fmt.Sprintf("field %s not found", order.Field))
			}

			v, err := breflect.GetStructField(item, fieldOrder.Name)
			if err != nil {
				return "", err
			}
			orderFieldValues[i] = v.Interface()
		}

		cursor := &model.Cursor{
			Value: orderFieldValues,
		}

		w := new(bytes.Buffer)
		err = cursor.Marshal(w)
		if err != nil {
			return "", err
		}

		return w.String(), nil
	}

	conn.Edges = make([]*model.Edge, breflect.SlicePtrLen(resultPtr))
	for i := range conn.Edges {
		node := nodeAt(i)

		var cursor string
		cursor, err = toCursor(node)
		if err != nil {
			return
		}

		conn.Edges[i] = &model.Edge{
			Node:   node,
			Cursor: cursor,
		}
	}

	conn.PageInfo.StartCursor = conn.Edges[0].Cursor
	conn.PageInfo.EndCursor = conn.Edges[len(conn.Edges)-1].Cursor

	return
}

func (p *connectionPaginator) applyCursor(queryHandler *_gorm.DB, query *model.ConnectionQuery) (*_gorm.DB, error) {
	if query.After != nil {
		if len(*query.After) != 0 {
			cursor := &model.Cursor{}
			err := cursor.Unmarshal(*query.After)
			if err != nil {
				return nil, err
			}

			if query.Orders != nil {
				if len(cursor.Value) != len(query.Orders) {
					return nil, errors.New(fmt.Sprintf("cursor format fields length: %d not match orders fields length: %d", len(cursor.Value), len(query.Orders)))
				}

				fields := make([]string, len(query.Orders))
				values := make([]interface{}, len(query.Orders))

				for i, order := range query.Orders {
					fieldOrder, ok := p.findField(order.Field, queryHandler)
					if !ok {
						err := errors.New(fmt.Sprintf("ERR_DB_UNKNOWN_FIELD %s", order.Field))
						return nil, err
					}
					fields[i] = order.Field

					switch fieldOrder.Struct.Type.String() {
					case "time.Time", "*time.Time":
						v, err := smarttime.Parse(cursor.Value[i])
						if err == nil {
							values[i] = time.Time(v)
						}
					default:
						values[i] = cursor.Value[i]
					}
				}

				// 以第一个排序字段的顺序为准, 合并比较
				if query.Orders[0].Direction != model.OrderDirectionDesc {
					queryHandler = queryHandler.Where(fmt.Sprintf("(%s) > (?)", strings.Join(fields, ",")), values)
				} else {
					queryHandler = queryHandler.Where(fmt.Sprintf("(%s) < (?)", strings.Join(fields, ",")), values)
				}
			}
		}
	}

	if query.Before != nil {
		if len(*query.Before) != 0 {
			cursor := &model.Cursor{}
			err := cursor.Unmarshal(*query.Before)
			if err != nil {
				return nil, err
			}

			if query.Orders != nil {
				if len(cursor.Value) != len(query.Orders) {
					return nil, errors.New(fmt.Sprintf("cursor format fields length: %d not match orders fields length: %d", len(cursor.Value), len(query.Orders)))
				}

				fields := make([]string, len(query.Orders))
				values := make([]interface{}, len(query.Orders))

				for i, order := range query.Orders {
					fieldOrder, ok := p.findField(order.Field, queryHandler)
					if !ok {
						err := errors.New(fmt.Sprintf("ERR_DB_UNKNOWN_FIELD %s", order.Field))
						return nil, err
					}
					fields[i] = order.Field

					switch fieldOrder.Struct.Type.String() {
					case "time.Time", "*time.Time":
						v, err := smarttime.Parse(cursor.Value[i])
						if err == nil {
							values[i] = time.Time(v)
						}
					default:
						values[i] = cursor.Value[i]
					}
				}

				if query.Orders[0].Direction != model.OrderDirectionDesc {
					queryHandler = queryHandler.Where(fmt.Sprintf("(%s) < (?)", strings.Join(fields, ",")), values)
				} else {
					queryHandler = queryHandler.Where(fmt.Sprintf("(%s) > (?)", strings.Join(fields, ",")), values)
				}
			}
		}
	}

	return queryHandler, nil
}

func (p *connectionPaginator) applyOrders(queryHandler *_gorm.DB, orders []*model.Order, reverse bool) (*_gorm.DB, error) {
	for _, order := range orders {
		if reverse {
			order.Direction = order.Direction.Reverse()
		}

		fieldOrder, ok := p.findField(order.Field, queryHandler)
		if !ok {
			return nil, errors.New(fmt.Sprintf("ERR_DB_UNKNOWN_FIELD %s", order.Field))
		}

		queryHandler = queryHandler.Order(fmt.Sprintf("%s %s", fieldOrder.DBName, order.Direction.String()))
	}

	return queryHandler, nil
}

// 约定 name为小写
func (p *connectionPaginator) findField(name string, dbHandler *_gorm.DB) (*_gorm.StructField, bool) {
	tableName := p.modelStruct.TableName(dbHandler)
	fieldsMap := fieldsCache[tableName]
	if fieldsMap == nil {
		fieldsMap = make(map[string]*_gorm.StructField)

		for _, field := range p.modelStruct.StructFields {
			fieldName := field.Tag.Get("json")
			fieldsMap[fieldName] = field
			logger.Infof("fieldName: %s", fieldName)
		}

		fieldsCache[tableName] = fieldsMap
	}
	field, ok := fieldsMap[name]
	return field, ok
}
