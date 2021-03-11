package mongo

import (
	"errors"
	"fmt"
	"time"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository/mongo/reflect"
	"github.com/duolacloud/microbase/types/smarttime"
	"gopkg.in/mgo.v2/bson"
)

func buildQuery(ms *reflect.StructInfo, filters map[string]interface{}) (bson.M, error) {
	bFilters := bson.M{}
	if filters == nil || len(filters) == 0 {
		return bFilters, nil
	}

	for k, v := range filters {
		filterType := entity.FilterType(k)

		switch filterType {
		case entity.FilterType_AND:
			subFilters, ok := v.([]interface{})
			if !ok {
				return nil, errors.New("ERR_MALFORMED_PARAMETERS")
			}

			subBFilters := make([]bson.M, len(subFilters))
			for i, sub := range subFilters {
				subFilter, ok := sub.(map[string]interface{})
				if !ok {
					return nil, errors.New("ERR_MALFORMED_PARAMETERS")
				}

				subBFilter, err := buildQuery(ms, subFilter)
				if err != nil {
					return nil, err
				}
				subBFilters[i] = subBFilter
			}
			bFilters["$and"] = subBFilters
		case entity.FilterType_OR:
			subFilters, ok := v.([]interface{})
			if !ok {
				return nil, errors.New("ERR_MALFORMED_PARAMETERS")
			}

			subBFilters := make([]bson.M, len(subFilters))
			for i, sub := range subFilters {
				subFilter, ok := sub.(map[string]interface{})
				if !ok {
					return nil, errors.New("ERR_MALFORMED_PARAMETERS")
				}

				subBFilter, err := buildQuery(ms, subFilter)
				if err != nil {
					return nil, err
				}
				subBFilters[i] = subBFilter
				bFilters["$or"] = subBFilters
			}
		case entity.FilterType_NOR:
			subFilters, ok := v.([]interface{})
			if !ok {
				return nil, errors.New("ERR_MALFORMED_PARAMETERS")
			}

			subBFilters := make([]bson.M, len(subFilters))
			for i, sub := range subFilters {
				subFilter, ok := sub.(map[string]interface{})
				if !ok {
					return nil, errors.New("ERR_MALFORMED_PARAMETERS")
				}

				subBFilter, err := buildQuery(ms, subFilter)
				if err != nil {
					return nil, err
				}
				subBFilters[i] = subBFilter
				bFilters["$nor"] = subBFilters
			}
		default:
			field, ok := ms.FieldsMap[k]
			if !ok {
				err := errors.New(fmt.Sprintf("ERR_DB_UNKNOWN_FIELD %s", k))
				return nil, err
			}

			bFilter, err := buildMongoFilter(field, v)
			if err != nil {
				return nil, err
			}
			bFilters[k] = bFilter
		}
	}
	return bFilters, nil
}

func buildMongoFilter(field *reflect.StructField, value interface{}) (bson.M, error) {
	vMap, ok := value.(map[string]interface{})
	if !ok {
		switch field.FieldType.String() {
		case "time.Time", "*time.Time":
			v, err := smarttime.Parse(value)
			if err == nil {
				value = time.Time(v)
			}
		}
		return bson.M{"$eq": value}, nil
	}

	for vKey, vValue := range vMap {
		switch field.FieldType.String() {
		case "time.Time", "*time.Time":
			v, err := smarttime.Parse(vValue)
			if err == nil {
				vValue = time.Time(v)
			}
		}

		filterType := entity.FilterType(vKey)
		switch filterType {
		case entity.FilterType_EQ:
			return bson.M{"$eq": vValue}, nil
		case entity.FilterType_NE:
			return bson.M{"$ne": vValue}, nil
		case entity.FilterType_GT:
			return bson.M{"$gt": vValue}, nil
		case entity.FilterType_GTE:
			return bson.M{"$gte": vValue}, nil
		case entity.FilterType_LT:
			return bson.M{"$lt": vValue}, nil
		case entity.FilterType_LTE:
			return bson.M{"$lte": vValue}, nil
		case entity.FilterType_LIKE:
			return bson.M{"$regex": vValue}, nil
		case entity.FilterType_MATCH:
			return bson.M{"$regex": vValue}, nil
		case entity.FilterType_NOT_LIKE:
			return bson.M{"$not": bson.M{"$regex": vValue}}, nil
		case entity.FilterType_IN:
			return bson.M{"$in": vValue}, nil
		case entity.FilterType_NOT_IN:
			return bson.M{"$nin": vValue}, nil
		case entity.FilterType_IS_NULL:
			return bson.M{"$exists": false}, nil
		case entity.FilterType_NOT_NULL:
			return bson.M{"$exists": true}, nil
		default:
			return nil, errors.New("ERR_MALFORMED_FILTER_TYPE")
		}
	}
	return bson.M{}, nil
}

func applyOrders(ms *reflect.StructInfo, sorts []*entity.Order) ([]string, error) {
	if orders == nil {
		return nil, nil
	}

	borders := make([]string{}, len(orders))

	for _, s := range orders {
		if _, ok := ms.FieldsMap[s.Property]; !ok {
			return nil, errors.New(fmt.Sprintf("ERR_DB_UNKNOWN_FIELD %s", s.Property))
		}
		var s1 string
		switch s.Type {
		case entity.OrderDirectionDesc:
			{
				s1 = fmt.Sprintf("-%s", s.Property)
			}
		default: // SortDirection_ASC
			{
				s1 = s.Property
			}
		}
		borders = append(borders, s1)
	}

	return borders, nil
}
