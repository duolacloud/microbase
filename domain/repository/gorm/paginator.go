package gorm

import (
	"context"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"
	_gorm "github.com/jinzhu/gorm"
)

type paginator struct {
	dataSourceProvider repository.DataSourceProvider
	entity             entity.Entity
	modelStruct        *_gorm.ModelStruct
}

func NewPaginator(dataSourceProvider repository.DataSourceProvider, entity entity.Entity) repository.Paginator {
	return &paginator{
		dataSourceProvider: dataSourceProvider,
		entity:             entity,
	}
}

func (p *paginator) DB(c context.Context) (*_gorm.DB, error) {
	db, err := p.dataSourceProvider.ProvideDB(c)
	if err != nil {
		return nil, err
	}
	return db.(*_gorm.DB), nil
}

func (p *paginator) Paginate(c context.Context, query *entity.PageQuery, resultPtr interface{}) (total int64, err error) {
	db, err := p.DB(c)
	if err != nil {
		return
	}

	scope := db.NewScope(p.entity)
	table := p.dataSourceProvider.ProvideTable(c, scope.TableName())

	if p.modelStruct == nil {
		p.modelStruct = db.NewScope(p.entity).GetModelStruct()
	}

	dbHandler := db.Table(table)
	dbHandler, err = applyFilter(dbHandler, p.modelStruct, query.Filter)
	if err != nil {
		return
	}

	dbHandler, err = applyOrders(dbHandler, p.modelStruct, query.Orders)
	if err != nil {
		return
	}

	total, err = pageQuery(dbHandler, query.PageNo, query.PageSize, resultPtr)
	return
}
