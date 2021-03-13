package elasticsearch

import (
	"context"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"
)

type Paginator struct {
	dataSourceProvider repository.DataSourceProvider
}

func NewPaginator(dataSourceProvider repository.DataSourceProvider) *Paginator {
	return &Paginator{
		dataSourceProvider,
	}
}

func (p *Paginator) Paginate(c context.Context, query *entity.PageQuery, index, typ string) (docs []*entity.Document, total int, pageCount int, err error) {
	return
}
