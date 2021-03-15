package elasticsearch

import (
	"context"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/olivere/elastic/v6"
)

type ConnectionPaginator struct {
	client *elastic.Client
}

func NewConnectionPaginator(client *elastic.Client) *ConnectionPaginator {
	return &ConnectionPaginator{
		client,
	}
}

func (p *ConnectionPaginator) Paginate(c context.Context, query *entity.ConnectionQuery, index, typ string) (conn *entity.Connection, err error) {
	return
}
