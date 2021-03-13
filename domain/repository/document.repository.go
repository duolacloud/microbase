package repository

import (
	"context"

	"github.com/duolacloud/microbase/domain/entity"
)

type DocumentRepository interface {
	Create(c context.Context, doc *entity.Document) error

	Upsert(c context.Context, doc *entity.Document) error

	Update(c context.Context, doc *entity.Document) error

	Get(c context.Context, id string, index, typ string) (*entity.Document, error)

	// 翻页查询
	// query: 查询条件
	// m: 数据指针，仅用于帮助推导数据类型
	Page(c context.Context, query *entity.PageQuery, index, typ string) (docs []*entity.Document, total int, pageCount int, err error)

	// 根据主键删除数据
	// m	数据对象
	Delete(c context.Context, id string, index, typ string) error

	// 游标查询
	// @c	上下文
	// @query	查询条件
	// resultPtr	返回数据的指针
	List(c context.Context, query *entity.CursorQuery, index, typ string) (docs []*entity.Document, cursor *entity.CursorExtra, err error)

	// GraphQL 游标查询
	// @c	上下文
	// @query	查询条件
	Connection(c context.Context, query *entity.ConnectionQuery, index, typ string) (*entity.Connection, error)
}
