package gorm

import (
	"context"
	"errors"
	"fmt"

	"github.com/duolacloud/microbase/datasource/gorm/opentracing"
	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"
	breflect "github.com/duolacloud/microbase/reflect"
	_gorm "github.com/jinzhu/gorm"
)

type BaseRepository struct {
	DataSourceProvider repository.DataSourceProvider
}

func NewBaseRepository(dataSourceProvider repository.DataSourceProvider) repository.BaseRepository {
	return &BaseRepository{dataSourceProvider}
}

func (r *BaseRepository) DB(c context.Context) (*_gorm.DB, error) {
	db, err := r.DataSourceProvider.ProvideDB(c)
	if err != nil {
		return nil, err
	}
	return db.(*_gorm.DB), nil
}

func (r *BaseRepository) Create(c context.Context, m entity.Entity) error {
	db, err := r.DB(c)
	if err != nil {
		return err
	}
	db = opentracing.SetSpanToGorm(c, db)

	scope := db.NewScope(m)
	table := r.DataSourceProvider.ProvideTable(c, scope.TableName())

	return db.Table(table).Create(m).Error
}

func (r *BaseRepository) Upsert(c context.Context, m entity.Entity) (*repository.ChangeInfo, error) {
	db, err := r.DB(c)
	if err != nil {
		return nil, err
	}

	db = opentracing.SetSpanToGorm(c, db)
	scope := db.NewScope(m)
	table := r.DataSourceProvider.ProvideTable(c, scope.TableName())

	result := db.Table(table).Save(m)
	if result.Error != nil {
		return nil, result.Error
	}

	change := &repository.ChangeInfo{
		Updated: int(result.RowsAffected),
	}
	return change, nil
}

func (r *BaseRepository) Update(c context.Context, m entity.Entity, data interface{}) error {
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

	table := r.DataSourceProvider.ProvideTable(c, scope.TableName())

	return db.Table(table).Where(m.Unique()).Update(data).Error
}

func (r *BaseRepository) Get(c context.Context, m entity.Entity) error {
	db, err := r.DB(c)
	if err != nil {
		return err
	}
	scope := db.NewScope(m)
	table := r.DataSourceProvider.ProvideTable(c, scope.TableName())

	db = opentracing.SetSpanToGorm(c, db)

	return db.Table(table).Where(m.Unique()).Take(m).Error
}

func (r *BaseRepository) Delete(c context.Context, m entity.Entity) error {
	db, err := r.DB(c)
	if err != nil {
		return err
	}
	scope := db.NewScope(m)
	table := r.DataSourceProvider.ProvideTable(c, scope.TableName())

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

	return db.Table(table).Delete(m).Error
}

func (r *BaseRepository) Page(c context.Context, m entity.Entity, query *entity.PageQuery, resultPtr interface{}) (total int, pageCount int, err error) {
	db, err := r.DB(c)
	if err != nil {
		return
	}
	scope := db.NewScope(m)
	table := r.DataSourceProvider.ProvideTable(c, scope.TableName())

	// items := breflect.MakeSlicePtr(m, 0, 0)
	ms := db.NewScope(m).GetModelStruct()

	dbHandler := db.Table(table)
	dbHandler, err = applyFilter(dbHandler, ms, query.Filter)
	if err != nil {
		return
	}

	dbHandler, err = applyOrders(dbHandler, ms, query.Orders)
	if err != nil {
		return
	}

	total, pageCount, err = pageQuery(dbHandler, query.PageNo, query.PageSize, resultPtr)

	return
}

func (r *BaseRepository) List(c context.Context, query *entity.CursorQuery, entity entity.Entity, resultPtr interface{}) (extra *entity.CursorExtra, err error) {
	paginator := NewCursorPaginator(r.DataSourceProvider, entity)

	extra, err = paginator.Paginate(c, query, resultPtr)

	return
}

func (r *BaseRepository) Connection(c context.Context, query *entity.ConnectionQuery, entity entity.Entity) (*entity.Connection, error) {
	paginator := NewConnectionPaginator(r.DataSourceProvider, entity)

	return paginator.Paginate(c, query)
}
