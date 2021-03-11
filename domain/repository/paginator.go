package repository

import (
	"context"

	"github.com/duolacloud/microbase/domain/model"
)

type ConnectionPaginator interface {
	Paginate(c context.Context, query *model.ConnectionQuery) (conn *model.Connection, err error)
}

type CursorPaginator interface {
	Paginate(c context.Context, query *model.CursorQuery, resultPtr interface{}) (extra *model.CursorExtra, err error)
}
