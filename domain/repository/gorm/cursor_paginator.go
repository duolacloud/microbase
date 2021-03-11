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

type cursorPaginator struct {
	db          *_gorm.DB
	model       model.Model
	modelStruct *_gorm.ModelStruct
}

func NewCursorPaginator(db *_gorm.DB, model model.Model) repository.CursorPaginator {
	modelStruct := db.NewScope(model).GetModelStruct()

	return &cursorPaginator{
		db,
		model,
		modelStruct,
	}
}

func (p *cursorPaginator) Paginate(c context.Context, query *model.CursorQuery, resultPtr interface{}) (extra *model.CursorExtra, err error) {
	if len(p.modelStruct.PrimaryFields) == 0 {
		err = errors.New("no primary key found for model")
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

	extra = &model.CursorExtra{}

	dbHandler := p.db.Model(p.model)
	dbHandler, err = applyFilter(dbHandler, p.modelStruct, query.Filter)
	if err != nil {
		return nil, err
	}

	if query.NeedTotal {
		if err = dbHandler.Count(&extra.Total).Error; err != nil {
			return
		}
	}

	dbHandler, err = p.applyCursor(dbHandler, query)
	if err != nil {
		return nil, err
	}

	dbHandler, err = p.applyOrders(dbHandler, query.Orders, query.Direction == model.CursorDirectionBefore)
	if err != nil {
		return nil, err
	}

	limit := query.Size + 1
	dbHandler = dbHandler.Limit(limit)

	if len(query.Fields) > 0 {
		dbHandler = dbHandler.Select(query.Fields)
	}

	if err = dbHandler.Find(resultPtr).Error; err != nil {
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

func (p *cursorPaginator) applyCursor(queryHandler *_gorm.DB, query *model.CursorQuery) (*_gorm.DB, error) {
	if len(query.Cursor) > 0 {
		cursor := &model.Cursor{}
		err := cursor.Unmarshal(query.Cursor)
		if err != nil {
			return nil, err
		}

		if len(cursor.Value) == 0 {
			return queryHandler, nil
		}

		if len(cursor.Value) != len(query.Orders) {
			return nil, errors.New(fmt.Sprintf("cursor format fields length: %d not match orders fields length: %d", len(cursor.Value), len(query.Orders)))
		}

		fields := make([]string, len(cursor.Value))
		values := make([]interface{}, len(cursor.Value))

		for i, value := range cursor.Value {
			order := query.Orders[i]

			fieldOrder, ok := p.findField(order.Field, queryHandler)
			if !ok {
				err := errors.New(fmt.Sprintf("ERR_DB_UNKNOWN_FIELD %s", order.Field))
				return nil, err
			}
			fields[i] = order.Field

			switch fieldOrder.Struct.Type.String() {
			case "time.Time", "*time.Time":
				v, err := smarttime.Parse(value)
				if err == nil {
					values[i] = time.Time(v)
				}
			default:
				values[i] = value
			}
		}

		if query.Direction == model.CursorDirectionAfter {
			// after
			if query.Orders[0].Direction == model.OrderDirectionDesc {
				queryHandler = queryHandler.Where(fmt.Sprintf("(%s) < (?)", strings.Join(fields, ",")), values)
			} else {
				queryHandler = queryHandler.Where(fmt.Sprintf("(%s) > (?)", strings.Join(fields, ",")), values)
			}
		} else {
			// before
			if query.Orders[0].Direction == model.OrderDirectionDesc {
				queryHandler = queryHandler.Where(fmt.Sprintf("(%s) > (?)", strings.Join(fields, ",")), values)
			} else {
				queryHandler = queryHandler.Where(fmt.Sprintf("(%s) < (?)", strings.Join(fields, ",")), values)
			}
		}
	}

	return queryHandler, nil
}

func (p *cursorPaginator) applyOrders(queryHandler *_gorm.DB, orders []*model.Order, reverse bool) (*_gorm.DB, error) {
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
func (p *cursorPaginator) findField(name string, dbHandler *_gorm.DB) (*_gorm.StructField, bool) {
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
