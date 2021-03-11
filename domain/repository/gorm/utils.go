package gorm

import (
	"errors"
	"fmt"
	"time"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/logger"
	"github.com/duolacloud/microbase/types/smarttime"
	_gorm "github.com/jinzhu/gorm"
)

var (
	fieldsCache = make(map[string]map[string]*_gorm.StructField)
)

// 约定 name为小写
func FindField(name string, ms *_gorm.ModelStruct, dbHandler *_gorm.DB) (*_gorm.StructField, bool) {
	tableName := ms.TableName(dbHandler)
	fieldsMap := fieldsCache[tableName]
	if fieldsMap == nil {
		fieldsMap = make(map[string]*_gorm.StructField)

		for _, field := range ms.StructFields {
			fieldName := field.Tag.Get("json")
			fieldsMap[fieldName] = field
			logger.Infof("fieldName: %s", fieldName)
		}

		fieldsCache[tableName] = fieldsMap
	}
	field, ok := fieldsMap[name]
	return field, ok
}

func applyFilter(db *_gorm.DB, ms *_gorm.ModelStruct, filters map[string]interface{}) (*_gorm.DB, error) {
	if filters == nil || len(filters) == 0 {
		return db, nil
	}

	var err error
	for key, value := range filters {
		db, err = gormFilter(db, ms, key, value)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func gormFilter(db *_gorm.DB, ms *_gorm.ModelStruct, key string, value interface{}) (*_gorm.DB, error) {
	filterType := entity.FilterType(key)

	switch filterType {
	case entity.FilterType_AND:
		{
			/* TODO 暂时默认就是 AND
			subFilters := v.([]interface{})
			for _, item := range subFilters {
				db = buildQuery(db, subFilter, ms)
			}*/
		}
	case entity.FilterType_OR:
		{
			/* TODO  暂时不支持 page 中支持 or
			for _, item := range subFilters {
				db := buildQuery(db, subFilter, ms)
				orCond = orCond.Or(subCond)
			}
			*/
		}
	default:
		{
			field, ok := FindField(key, ms, db)
			if !ok {
				err := errors.New(fmt.Sprintf("ERR_DB_UNKNOWN_FIELD %s", key))
				return nil, err
			}

			fieldName := field.DBName

			vMap, ok := value.(map[string]interface{})
			if !ok {
				switch field.Struct.Type.String() {
				case "time.Time", "*time.Time":
					v, err := smarttime.Parse(value)
					if err == nil {
						value = time.Time(v)
					}
				}

				return db.Where(fmt.Sprintf("%s = ?", fieldName), value), nil
			}

			for vKey, vValue := range vMap {
				switch field.Struct.Type.String() {
				case "time.Time", "*time.Time":
					v, err := smarttime.Parse(vValue)
					if err == nil {
						vValue = v
					}
				}

				filterType = entity.FilterType(vKey)
				switch filterType {
				case entity.FilterType_EQ:
					return db.Where(fmt.Sprintf("%s = ?", fieldName), vValue), nil
				case entity.FilterType_NE:
					return db.Where(fmt.Sprintf("%s != ?", fieldName), vValue), nil
				case entity.FilterType_GT:
					return db.Where(fmt.Sprintf("%s > ?", fieldName), vValue), nil
				case entity.FilterType_GTE:
					return db.Where(fmt.Sprintf("%s >= ?", fieldName), vValue), nil
				case entity.FilterType_LT:
					return db.Where(fmt.Sprintf("%s < ?", fieldName), vValue), nil
				case entity.FilterType_LTE:
					return db.Where(fmt.Sprintf("%s <= ?", fieldName), vValue), nil
				case entity.FilterType_LIKE:
					return db.Where(fmt.Sprintf("%s LIKE ?", fieldName), vValue), nil
				case entity.FilterType_MATCH:
					return db.Where(fmt.Sprintf("%s LIKE ?", fieldName), vValue), nil
				case entity.FilterType_NOT_LIKE:
					return db.Not(fmt.Sprintf("%s LIKE ?", fieldName), vValue), nil
				case entity.FilterType_IN:
					return gormFilterIn(db, fieldName, vValue)
				case entity.FilterType_NOT_IN:
					return gormFilterNotIn(db, fieldName, vValue)
				case entity.FilterType_BETWEEN:
					return gormFilterBetween(db, fieldName, vValue)
				case entity.FilterType_IS_NULL:
					return db.Where(fmt.Sprintf("%s IS NULL", fieldName)), nil
				case entity.FilterType_NOT_NULL:
					return db.Where(fmt.Sprintf("%s IS NOT NULL", fieldName)), nil
				}
			}
		}
	}

	return db, nil
}

func gormFilterIn(db *_gorm.DB, key string, value interface{}) (*_gorm.DB, error) {
	values, ok := value.([]interface{})
	if !ok {
		return nil, ErrFilterValueType
	}

	return db.Where(fmt.Sprintf("%s IN (?)", key), values), nil
}

func gormFilterNotIn(db *_gorm.DB, key string, value interface{}) (*_gorm.DB, error) {
	values, ok := value.([]interface{})
	if !ok {
		return nil, ErrFilterValueType
	}

	return db.Where(fmt.Sprintf("%s NOT IN (?)", key), values), nil
}

func gormFilterBetween(db *_gorm.DB, key string, value interface{}) (*_gorm.DB, error) {
	values, ok := value.([]interface{})
	if !ok {
		return nil, ErrFilterValueType
	}
	if len(values) != 2 {
		return nil, ErrFilterValueSize
	}
	if values[0] != nil && values[1] != nil {
		return db.Where(fmt.Sprintf("%s between ? and ?", key), values[0], values[1]), nil
	} else if values[0] != nil && values[1] == nil {
		return db.Where(fmt.Sprintf("%s >= ?", key), values[0]), nil
	} else if values[0] == nil && values[1] != nil {
		return db.Where(fmt.Sprintf("%s <= ?", key), values[1]), nil
	} else {
		return db, nil
	}
}

func applyOrders(dbHandler *_gorm.DB, ms *_gorm.ModelStruct, orders []*entity.Order) (*_gorm.DB, error) {
	if orders == nil || len(orders) == 0 {
		return dbHandler, nil
	}

	for _, order := range orders {
		field, ok := FindField(order.Field, ms, dbHandler)
		if !ok {
			return nil, errors.New(fmt.Sprintf("unknown field: %s", order.Field))
		}

		dbHandler = dbHandler.Order(fmt.Sprintf("%s %s", field.DBName, order.Direction.String()))
	}

	return dbHandler, nil
}
