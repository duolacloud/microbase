/**
 * Facebook 的 Graphql relay 查询模式
 */
package gorm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duolacloud/microbase/domain/model"
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

func NewConnectionPaginator(db *_gorm.DB, model model.Model) *repository.ConnectionPaginator {
	modelStruct := db.NewScope(model).GetModelStruct()

	return &connectionPaginator{
		db,
		model,
		modelStruct,
	}
}

func (p *connectionPaginator) Paginate(c context.Context, query *model.ConnectionQuery) (conn *model.Connection, err error) {
	if len(p.modelStruct.PrimaryFields) == 0 {
		err = errors.New("no primary key found for model")
		return
	}

	if len(p.modelStruct.PrimaryFields) != 1 {
		err = errors.New("more then one primary key found for model")
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

	dbHandler, err = p.applyCursor(dbHandler, query)
	if err != nil {
		return nil, err
	}

	dbHandler, err = p.applyOrders(dbHandler, query, query.Direction != 0)
	if err != nil {
		return nil, err
	}

	extra = &model.CursorExtra{}

	if query.NeedTotal {
		dbHandler.Count(&extra.Total)
	}

	limit := query.Size + 1
	if err = dbHandler.Limit(limit).Find(resultPtr).Error; err != nil {
		return
	}

	count := breflect.SlicePtrLen(resultPtr)
	if count == 0 {
		return
	}

	if count == limit {
		extra.HasNext = true
		extra.HasPrevious = true
		breflect.SlicePtrSlice3To(resultPtr, 0, limit-1, limit-1, resultPtr)
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

	itemCount := breflect.SlicePtrLen(resultPtr)
	extra.StartCursor, err = toCursor(breflect.SlicePtrIndexOf(resultPtr, 0))
	if err != nil {
		return
	}

	extra.EndCursor, err = toCursor(breflect.SlicePtrIndexOf(resultPtr, itemCount-1))
	if err != nil {
		return
	}

	return
}

func (p *CursorPaginator) applyCursor(queryHandler *_gorm.DB, query *model.CursorQuery) (*_gorm.DB, error) {
	if len(query.Cursor) > 0 {
		cursor := &model.Cursor{}
		err := cursor.Unmarshal(query.Cursor)
		if err != nil {
			return nil, err
		}

		// idField := p.modelStruct.PrimaryFields[0]

		if query.Orders != nil {
			if len(cursor.Value) != len(query.Orders) {
				return nil, errors.New(fmt.Sprintf("cursor format fields length: %d not match orders fields length: %d", len(cursor.Value), len(query.Orders)))
			}

			for i, order := range query.Orders {
				fieldOrder, ok := p.findField(order.Field, queryHandler)
				if !ok {
					err := errors.New(fmt.Sprintf("ERR_DB_UNKNOWN_FIELD %s", order.Field))
					return nil, err
				}

				var cursorValue interface{}

				switch fieldOrder.Struct.Type.String() {
				case "time.Time", "*time.Time":
					v, err := smarttime.Parse(cursor.Value[i])
					if err == nil {
						cursorValue = time.Time(v)
					}
				default:
					cursorValue = cursor.Value[i]
				}

				if query.Direction == 0 {
					queryHandler = queryHandler.Where(fmt.Sprintf("%s > ?", order.Field), cursorValue)
					// queryHandler = queryHandler.Where(fmt.Sprintf("%s > ?", idField.DBName), cursor.ID)
				} else {
					queryHandler = queryHandler.Where(fmt.Sprintf("%s < ?", order.Field), cursorValue)
					// queryHandler = queryHandler.Where(fmt.Sprintf("%s < ?", idField.DBName), cursor.ID)
				}
			}
		}
	}

	return queryHandler, nil
}

func (p *CursorPaginator) applyOrders(queryHandler *_gorm.DB, query *model.CursorQuery, reverse bool) (*_gorm.DB, error) {
	for _, order := range query.Orders {
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
func (p *CursorPaginator) findField(name string, dbHandler *_gorm.DB) (*_gorm.StructField, bool) {
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
