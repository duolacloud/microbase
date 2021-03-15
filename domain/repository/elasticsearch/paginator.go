package elasticsearch

import (
	"context"

	"github.com/duolacloud/microbase/client/search"
	"github.com/duolacloud/microbase/domain/entity"
	"github.com/olivere/elastic/v6"
)

type Paginator struct {
	client *elastic.Client
}

func NewPaginator(client *elastic.Client) *Paginator {
	return &Paginator{
		client,
	}
}

func (p *Paginator) Paginate(c context.Context, query *entity.PageQuery, index, typ string) (docs []*search.Document, total int64, err error) {
	return
}
