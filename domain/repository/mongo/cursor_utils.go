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

func mongoCursorFilter(ms *reflect.StructInfo, cursorQuery *entity.CursorQuery) (filter bson.M, sort string, reverse bool, err error) {
	prop := cursorQuery.CursorSort.Property
	field, ok := ms.FieldsMap[prop]
	if !ok {
		err = errors.New(fmt.Sprintf("ERR_DB_UNKNOWN_FIELD %s", prop))
		return
	}

	value := cursorQuery.Cursor
	switch field.FieldType.String() {
	case "time.Time", "*time.Time":
		v, err := smarttime.Parse(value)
		if err == nil {
			value = time.Time(v)
		}
	}

	switch cursorQuery.CursorSort.Type {
	case entity.SortDirection_DSC:
		{
			if cursorQuery.Direction == 0 {
				// 游标前
				sort = prop
				reverse = true
				if value != nil {
					filter = bson.M{prop: bson.M{"$gt": value}}
				}
			} else {
				// 游标后
				sort = fmt.Sprintf("-%s", prop)
				reverse = false
				if value != nil {
					filter = bson.M{prop: bson.M{"$lt": value}}
				}
			}
		}
	default: // SortDirection_ASC
		{
			if cursorQuery.Direction == 0 {
				// 游标前
				sort = fmt.Sprintf("-%s", prop)
				reverse = true
				if value != nil {
					filter = bson.M{prop: bson.M{"$lt": value}}
				}
			} else {
				// 游标后
				sort = prop
				reverse = false
				if value != nil {
					filter = bson.M{prop: bson.M{"$gt": value}}
				}
			}
		}
	}

	return
}
