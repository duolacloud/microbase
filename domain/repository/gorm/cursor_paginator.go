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
	"github.com/duolacloud/microbase/logger"
	breflect "github.com/duolacloud/microbase/reflect"
	"github.com/duolacloud/microbase/types/smarttime"
	_gorm "github.com/jinzhu/gorm"
)

type cursorPaginator struct {
	dataSourceProvider repository.DataSourceProvider
	entity             entity.Entity
	modelStruct        *_gorm.ModelStruct
}

func NewCursorPaginator(dataSourceProvider repository.DataSourceProvider, entity entity.Entity) repository.CursorPaginator {
	return &cursorPaginator{
		dataSourceProvider: dataSourceProvider,
		entity:             entity,
	}
}

func (p *cursorPaginator) Paginate(c context.Context, query *entity.CursorQuery, resultPtr interface{}) (extra *entity.CursorExtra, err error) {
	_db, err := p.dataSourceProvider.ProvideDB(c)
	if err != nil {
		return nil, err
	}
	db := _db.(*_gorm.DB)

	scope := db.NewScope(p.entity)
	table := p.dataSourceProvider.ProvideTable(c, scope.TableName())

	if p.modelStruct == nil {
		p.modelStruct = db.NewScope(p.entity).GetModelStruct()
	}

	if len(p.modelStruct.PrimaryFields) == 0 {
		err = errors.New("no primary key found for entity")
		return
	}

	if query.Orders == nil {
		fieldIDs := p.modelStruct.PrimaryFields
		query.Orders = make([]*entity.Order, len(fieldIDs))
		for i, fieldID := range fieldIDs {
			order := &entity.Order{
				Field:     fieldID.DBName,
				Direction: entity.OrderDirectionAsc,
			}
			query.Orders[i] = order
		}
	}

	p.ensureOrders(p.modelStruct, query)

	extra = &entity.CursorExtra{}

	dbHandler := db.Table(table)
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

	dbHandler, err = p.applyOrders(dbHandler, query.Orders, query.Direction == entity.CursorDirectionBefore)
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

func (p *cursorPaginator) ensureOrders(modelStruct *_gorm.ModelStruct, query *entity.CursorQuery) {
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
			query.Orders = make([]*entity.Order, len(fieldIDs))
		}

		for i, fieldID := range fieldIDs {
			order := &entity.Order{
				Field:     fieldID.DBName,
				Direction: entity.OrderDirectionAsc,
			}
			query.Orders[i] = order
		}
	}
}

func (p *cursorPaginator) applyCursor(queryHandler *_gorm.DB, query *entity.CursorQuery) (*_gorm.DB, error) {
	if len(query.Cursor) > 0 {
		cursor := &entity.Cursor{}
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

		if query.Direction == entity.CursorDirectionAfter {
			// after
			if query.Orders[0].Direction == entity.OrderDirectionDesc {
				queryHandler = queryHandler.Where(fmt.Sprintf("(%s) < (?)", strings.Join(fields, ",")), values)
			} else {
				queryHandler = queryHandler.Where(fmt.Sprintf("(%s) > (?)", strings.Join(fields, ",")), values)
			}
		} else {
			// before
			if query.Orders[0].Direction == entity.OrderDirectionDesc {
				queryHandler = queryHandler.Where(fmt.Sprintf("(%s) > (?)", strings.Join(fields, ",")), values)
			} else {
				queryHandler = queryHandler.Where(fmt.Sprintf("(%s) < (?)", strings.Join(fields, ",")), values)
			}
		}
	}

	return queryHandler, nil
}

func (p *cursorPaginator) applyOrders(queryHandler *_gorm.DB, orders []*entity.Order, reverse bool) (*_gorm.DB, error) {
	for _, order := range orders {
		if reverse {
			order.Direction = orders[0].Direction.Reverse()
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
