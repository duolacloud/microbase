package repository

import (
	"context"

	"github.com/duolacloud/microbase/domain/entity"
)

type Paginator interface {
	Paginate(c context.Context, query *entity.PageQuery, resultPtr interface{}) (total int, pageCount int, err error)
}

type ConnectionPaginator interface {
	Paginate(c context.Context, query *entity.ConnectionQuery) (conn *entity.Connection, err error)
}

type CursorPaginator interface {
	Paginate(c context.Context, query *entity.CursorQuery, resultPtr interface{}) (extra *entity.CursorExtra, err error)
}
