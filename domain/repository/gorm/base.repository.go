package gorm

import (
	"context"
	"errors"
	"fmt"

	"github.com/duolacloud/microbase/database/gorm/opentracing"
	"github.com/duolacloud/microbase/domain/model"
	"github.com/duolacloud/microbase/domain/repository"
	breflect "github.com/duolacloud/microbase/reflect"
	_gorm "github.com/jinzhu/gorm"
)

type BaseRepository struct {
	DBProvider repository.DBProvider
}

func NewBaseRepository(provider repository.DBProvider) repository.BaseRepository {
	return &BaseRepository{provider}
}

func (r *BaseRepository) DB(c context.Context) (*_gorm.DB, error) {
	db, err := r.DBProvider.Provide(c)
	if err != nil {
		return nil, err
	}
	return db.(*_gorm.DB), nil
}

func (r *BaseRepository) Create(c context.Context, m model.Model) error {
	db, err := r.DB(c)
	if err != nil {
		return err
	}

	db = opentracing.SetSpanToGorm(c, db)

	return db.Create(m).Error
}

func (r *BaseRepository) Upsert(c context.Context, m model.Model) (*repository.ChangeInfo, error) {
	db, err := r.DB(c)
	if err != nil {
		return nil, err
	}

	db = opentracing.SetSpanToGorm(c, db)

	result := db.Save(m)
	if result.Error != nil {
		return nil, result.Error
	}

	change := &repository.ChangeInfo{
		Updated: int(result.RowsAffected),
	}
	return change, nil
}

func (r *BaseRepository) Update(c context.Context, m model.Model, data interface{}) error {
	db, err := r.DB(c)
	if err != nil {
		return err
	}

	db = opentracing.SetSpanToGorm(c, db)

	// 主键保护，如果 m 什么都没设置，这里将会删除表的所有记录
	scope := db.NewScope(m)
	if scope.PrimaryKeyZero() {
		return errors.New(fmt.Sprintf("primary key(%s) must be set for update", scope.PrimaryKey()))
	}

	return db.Model(m).Update(data).Error
}

func (r *BaseRepository) Get(c context.Context, m model.Model) error {
	db, err := r.DB(c)
	if err != nil {
		return err
	}

	db = opentracing.SetSpanToGorm(c, db)

	return db.Where(m.Unique()).Take(m).Error
}

func (r *BaseRepository) Delete(c context.Context, m model.Model) error {
	db, err := r.DB(c)
	if err != nil {
		return err
	}

	// 主键保护，如果 m 什么都没设置，这里将会删除表的所有记录
	ms := db.NewScope(m).GetModelStruct()
	for _, pf := range ms.PrimaryFields {
		value, err := breflect.GetStructField(m, pf.Name)
		if err != nil {
			return err
		}

		if breflect.IsBlank(value) {
			return errors.New(fmt.Sprintf("primary key %s must set for delete", pf.Name))
		}
	}

	return db.Delete(m).Error
}

func (r *BaseRepository) Page(c context.Context, m model.Model, query *model.PageQuery, resultPtr interface{}) (total int, pageCount int, err error) {
	db, err := r.DB(c)
	if err != nil {
		return
	}

	// items := breflect.MakeSlicePtr(m, 0, 0)
	ms := db.NewScope(m).GetModelStruct()

	dbHandler := db.Model(m)
	dbHandler, err = buildQuery(dbHandler, ms, query.Filters)
	if err != nil {
		return
	}

	dbHandler, err = buildSort(dbHandler, ms, query.Sort)
	if err != nil {
		return
	}

	total, pageCount, err = pageQuery(dbHandler, query.PageNo, query.PageSize, resultPtr)

	return
}

func (r *BaseRepository) List(c context.Context, query *model.CursorQuery, m model.Model, resultPtr interface{}) (extra *model.CursorExtra, err error) {
	db, err := r.DB(c)
	if err != nil {
		return
	}

	ms := db.NewScope(m).GetModelStruct()

	dbHandler := db.Model(m)
	dbHandler, err = buildQuery(dbHandler, ms, query.Filter)
	if err != nil {
		return
	}

	dbHandler, reverse, err := gormCursorFilter(dbHandler, ms, query)
	if err != nil {
		return
	}

	// items := breflect.MakeSlicePtr(m, 0, 0)

	var total int
	dbHandler.Count(&total)

	if err = dbHandler.Limit(query.Size).Find(resultPtr).Error; err != nil {
		return
	}

	if reverse {
		breflect.SlicePtrReverse(resultPtr)
	}

	var startCursor interface{} = nil
	var endCursor interface{} = nil

	count := breflect.SlicePtrLen(resultPtr)
	if count > 0 {
		minItem := breflect.SlicePtrIndexOf(resultPtr, 0)
		field, ok := FindField(query.CursorSort.Field, ms, dbHandler)
		if !ok {
			err = errors.New("field not found")
			return
		}

		startCursor, err = breflect.GetStructField(minItem, field.Name)
		if err != nil {
			return
		}

		maxItem := breflect.SlicePtrIndexOf(resultPtr, count-1)
		endCursor, err = breflect.GetStructField(maxItem, field.Name)
		if err != nil {
			return
		}
	}

	var hasPrevious bool
	var hasNext bool
	if query.Direction == model.Direction_ASC {
		hasNext = count == query.Size
	} else if query.Direction == model.Direction_DSC {
		hasPrevious = count == query.Size
	}

	extra = &model.CursorExtra{
		Direction:   query.Direction,
		Total:       total,
		HasPrevious: hasPrevious,
		HasNext:     hasNext,
		StartCursor: startCursor,
		EndCursor:   endCursor,
	}

	return
}
