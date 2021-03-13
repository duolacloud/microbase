package elasticsearch

import (
	"context"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"
)

type ConnectionPaginator struct {
	dataSourceProvider repository.DataSourceProvider
}

func NewConnectionPaginator(dataSourceProvider repository.DataSourceProvider) *ConnectionPaginator {
	return &ConnectionPaginator{
		dataSourceProvider,
	}
}

func (p *ConnectionPaginator) Paginate(c context.Context, query *entity.ConnectionQuery, index, typ string) (conn *entity.Connection, err error) {
	return
}
