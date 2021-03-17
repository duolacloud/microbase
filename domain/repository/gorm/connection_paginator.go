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

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"
	breflect "github.com/duolacloud/microbase/reflect"
	"github.com/duolacloud/microbase/types/smarttime"
	_gorm "github.com/jinzhu/gorm"
)

type connectionPaginator struct {
	dataSourceProvider repository.DataSourceProvider
	entity             entity.Entity
}

func NewConnectionPaginator(dataSourceProvider repository.DataSourceProvider, entity entity.Entity) repository.ConnectionPaginator {
	return &connectionPaginator{
		dataSourceProvider: dataSourceProvider,
		entity:             entity,
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

func (p *connectionPaginator) Paginate(c context.Context, query *entity.ConnectionQuery) (conn *entity.Connection, err error) {
	_db, err := p.dataSourceProvider.ProvideDB(c)
	if err != nil {
		return nil, err
	}
	db := _db.(*_gorm.DB)

	scope := db.NewScope(p.entity)
	modelStruct := scope.GetModelStruct()

	table := p.dataSourceProvider.ProvideTable(c, scope.TableName())

	if len(modelStruct.PrimaryFields) == 0 {
		err = errors.New("no primary key found for model")
		return
	}

	err = validateFirstLast(query.First, query.Last)
	if err != nil {
		return
	}

	p.ensureOrders(modelStruct, query)

	dbHandler := db.Table(table)
	dbHandler, err = applyFilter(dbHandler, modelStruct, query.Filter)
	if err != nil {
		return nil, err
	}

	conn = &entity.Connection{Edges: []*entity.Edge{}}
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

	dbHandler, err = p.applyCursor(dbHandler, query, modelStruct)
	if err != nil {
		return nil, err
	}

	dbHandler, err = p.applyOrders(dbHandler, query.Orders, query.Last != nil, modelStruct)
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

	resultPtr := breflect.MakeSlicePtr(p.entity, 0, limit)
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
			fieldOrder, ok := FindField(order.Field, modelStruct, dbHandler)
			if !ok {
				return "", errors.New(fmt.Sprintf("field %s not found", order.Field))
			}

			v, err := breflect.GetStructField(item, fieldOrder.Name)
			if err != nil {
				return "", err
			}
			orderFieldValues[i] = v.Interface()
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

	conn.Edges = make([]*entity.Edge, breflect.SlicePtrLen(resultPtr))
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

func (p *connectionPaginator) applyCursor(queryHandler *_gorm.DB, query *entity.ConnectionQuery, modelStruct *_gorm.ModelStruct) (*_gorm.DB, error) {
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

				fields := make([]string, len(query.Orders))
				values := make([]interface{}, len(query.Orders))

				for i, order := range query.Orders {
					fieldOrder, ok := FindField(order.Field, modelStruct, queryHandler)
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
				if query.Orders[0].Direction != entity.OrderDirectionDesc {
					queryHandler = queryHandler.Where(fmt.Sprintf("(%s) > (?)", strings.Join(fields, ",")), values)
				} else {
					queryHandler = queryHandler.Where(fmt.Sprintf("(%s) < (?)", strings.Join(fields, ",")), values)
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

				fields := make([]string, len(query.Orders))
				values := make([]interface{}, len(query.Orders))

				for i, order := range query.Orders {
					fieldOrder, ok := FindField(order.Field, modelStruct, queryHandler)
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

				if query.Orders[0].Direction != entity.OrderDirectionDesc {
					queryHandler = queryHandler.Where(fmt.Sprintf("(%s) < (?)", strings.Join(fields, ",")), values)
				} else {
					queryHandler = queryHandler.Where(fmt.Sprintf("(%s) > (?)", strings.Join(fields, ",")), values)
				}
			}
		}
	}

	return queryHandler, nil
}

func (p *connectionPaginator) ensureOrders(modelStruct *_gorm.ModelStruct, query *entity.ConnectionQuery) {
	fieldIDs := modelStruct.PrimaryFields

	// 排序一定要包含ID, 否则会遍历不完整
	matchCount := 0
	for _, fieldID := range fieldIDs {
		for _, order := range query.Orders {
			if order.Field == fieldID.DBName {
				matchCount += 1
			}
		}
	}

	if matchCount != len(fieldIDs) {
		if query.Orders == nil {
			query.Orders = make([]*entity.Order, 0)
		}

		for _, fieldID := range fieldIDs {
			order := &entity.Order{
				Field:     fieldID.DBName,
				Direction: entity.OrderDirectionAsc,
			}
			query.Orders = append(query.Orders, order)
		}
	}
}

func (p *connectionPaginator) applyOrders(queryHandler *_gorm.DB, orders []*entity.Order, reverse bool, modelStruct *_gorm.ModelStruct) (*_gorm.DB, error) {
	for _, order := range orders {
		if reverse {
			// 默认用第一个的方向
			order.Direction = orders[0].Direction.Reverse()
		}

		fieldOrder, ok := FindField(order.Field, modelStruct, queryHandler)
		if !ok {
			return nil, errors.New(fmt.Sprintf("ERR_DB_UNKNOWN_FIELD %s", order.Field))
		}

		queryHandler = queryHandler.Order(fmt.Sprintf("%s %s", fieldOrder.DBName, order.Direction.String()))
	}

	return queryHandler, nil
}
