package repositories

import (
	"context"

	"github.com/duolacloud/microbase/client/search"
	"github.com/duolacloud/microbase/domain/entity"
)

type IndexRepository interface {
	Create(c context.Context, index *search.Index) error
	Delete(c context.Context, index string) error
	IndexExists(c context.Context, index string) (bool, error)
}

type DocumentRepository interface {
	Create(c context.Context, doc *search.Document) error

	Upsert(c context.Context, doc *search.Document) error

	Update(c context.Context, doc *search.Document) error

	Get(c context.Context, index, typ, id string) (*search.Document, error)

	// 翻页查询
	// query: 查询条件
	// m: 数据指针，仅用于帮助推导数据类型
	Page(c context.Context, query *entity.PageQuery, index, typ string) (docs []*search.Document, total int64, err error)

	// 根据主键删除数据
	// m	数据对象
	Delete(c context.Context, index, typ, id string) error

	// 游标查询
	// @c	上下文
	// @query	查询条件
	// resultPtr	返回数据的指针
	List(c context.Context, query *entity.CursorQuery, index, typ string) (docs []*search.Document, cursor *entity.CursorExtra, err error)

	// GraphQL 游标查询
	// @c	上下文
	// @query	查询条件
	Connection(c context.Context, query *entity.ConnectionQuery, index, typ string) (*entity.Connection, error)
}
