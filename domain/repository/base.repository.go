package repository

import (
	"context"

	"github.com/duolacloud/microbase/domain/entity"
)

type ChangeInfo struct {
	Updated    int
	Removed    int         // Number of documents removed
	Matched    int         // Number of documents matched but not necessarily changed
	UpsertedId interface{} // Upserted _id field, when not explicitly provided
}

type BaseRepository interface {
	Create(c context.Context, m entity.Entity) error

	Upsert(c context.Context, m entity.Entity) (*ChangeInfo, error)

	Update(c context.Context, m entity.Entity, change interface{}) error

	Get(c context.Context, m entity.Entity) error

	// 翻页查询
	// query: 查询条件
	// m: 数据指针，仅用于帮助推导数据类型
	Page(c context.Context, m entity.Entity, query *entity.PageQuery, resultPtr interface{}) (total int64, err error)

	// 根据主键删除数据
	// m	数据对象
	Delete(c context.Context, m entity.Entity) error

	// 游标查询
	// @c	上下文
	// @query	查询条件
	// m	数据指针，仅用于帮助推导数据类型
	// resultPtr	返回数据的指针
	List(c context.Context, query *entity.CursorQuery, m entity.Entity, resultPtr interface{}) (cursor *entity.CursorExtra, err error)

	// GraphQL 游标查询
	// @c	上下文
	// @query	查询条件
	// m	数据指针，仅用于帮助推导数据类型
	Connection(c context.Context, query *entity.ConnectionQuery, m entity.Entity) (*entity.Connection, error)
}
