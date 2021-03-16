package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/duolacloud/microbase/client/search"
	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"
	breflect "github.com/duolacloud/microbase/reflect"

	"github.com/thoas/go-funk"
)

type BaseRepository struct {
	DataSourceProvider repository.DataSourceProvider
	SearchClient       search.SearchClient
}

func NewBaseRepository(dataSourceProvider repository.DataSourceProvider) repository.BaseRepository {
	return &BaseRepository{
		DataSourceProvider: dataSourceProvider,
	}
}

func (r *BaseRepository) Client(c context.Context) (search.SearchClient, error) {
	o, err := r.DataSourceProvider.ProvideDB(c)
	if err != nil {
		return nil, err
	}

	return o.(search.SearchClient), nil
}

func (r *BaseRepository) Create(c context.Context, ent entity.Entity) error {
	searchClient, err := r.Client(c)
	if err != nil {
		return err
	}

	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return err
	}
	typ := breflect.TheNamingStrategy.Table(ms.Name)
	index := r.DataSourceProvider.ProvideTable(c, typ)

	buf, err := json.Marshal(ent)
	if err != nil {
		return err
	}

	var fields map[string]interface{}
	err = json.Unmarshal(buf, &fields)
	if err != nil {
		return err
	}

	return searchClient.Create(c, &search.Document{
		Index:  index,
		Type:   typ,
		Fields: fields,
	})
}

func (r *BaseRepository) Upsert(c context.Context, ent entity.Entity) (*repository.ChangeInfo, error) {
	searchClient, err := r.Client(c)
	if err != nil {
		return nil, err
	}

	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return nil, err
	}
	typ := breflect.TheNamingStrategy.Table(ms.Name)
	index := r.DataSourceProvider.ProvideTable(c, typ)

	buf, err := json.Marshal(ent)
	if err != nil {
		return nil, err
	}

	var fields map[string]interface{}
	err = json.Unmarshal(buf, &fields)
	if err != nil {
		return nil, err
	}

	err = searchClient.Upsert(c, &search.Document{
		Index:  index,
		Type:   typ,
		Fields: fields,
	})

	if err != nil {
		return nil, err
	}

	return &repository.ChangeInfo{
		Updated: 1,
	}, nil
}

func (r *BaseRepository) Update(c context.Context, ent entity.Entity, data interface{}) error {
	searchClient, err := r.Client(c)
	if err != nil {
		return err
	}

	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return err
	}
	typ := breflect.TheNamingStrategy.Table(ms.Name)
	index := r.DataSourceProvider.ProvideTable(c, typ)

	buf, err := json.Marshal(ent)
	if err != nil {
		return err
	}

	var fields map[string]interface{}
	err = json.Unmarshal(buf, &fields)
	if err != nil {
		return err
	}

	return searchClient.Update(c, &search.Document{
		Index:  index,
		Type:   typ,
		Fields: fields,
	})
}

func (r *BaseRepository) Get(c context.Context, ent entity.Entity) error {
	searchClient, err := r.Client(c)
	if err != nil {
		return err
	}

	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return err
	}
	typ := breflect.TheNamingStrategy.Table(ms.Name)
	index := r.DataSourceProvider.ProvideTable(c, typ)

	// 命名约定，必须有 id 字段
	id, ok := funk.Get(ent, "ID").(string)
	if !ok {
		return errors.New(fmt.Sprintf("no id field for entity %v", ent))
	}

	doc, err := searchClient.Get(c, index, typ, id)
	if err != nil {
		return err
	}

	if doc == nil {
		return nil
	}

	b, err := json.Marshal(doc.Fields)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &ent)
	if err != nil {
		return err
	}

	return nil
}

func (r *BaseRepository) Delete(c context.Context, ent entity.Entity) error {
	searchClient, err := r.Client(c)
	if err != nil {
		return err
	}

	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return err
	}
	typ := breflect.TheNamingStrategy.Table(ms.Name)
	index := r.DataSourceProvider.ProvideTable(c, typ)

	// 命名约定，必须有 id 字段
	id, ok := funk.Get(ent, "id").(string)
	if !ok {
		return errors.New(fmt.Sprintf("no id field for entity %v", ent))
	}

	return searchClient.Delete(c, index, typ, id)
}

func (r *BaseRepository) Page(c context.Context, ent entity.Entity, query *entity.PageQuery, resultPtr interface{}) (total int64, err error) {
	searchClient, err1 := r.Client(c)
	if err1 != nil {
		err = err1
		return
	}

	ms, err1 := breflect.GetStructInfo(ent, nil)
	if err1 != nil {
		err = err1
		return
	}

	typ := breflect.TheNamingStrategy.Table(ms.Name)
	index := r.DataSourceProvider.ProvideTable(c, typ)

	var docs []*search.Document
	docs, total, err = searchClient.Page(c, query, index, typ)
	if err != nil {
		return
	}

	results := reflect.ValueOf(resultPtr).Elem()
	resultType := results.Type().Elem()
	results.Set(reflect.MakeSlice(results.Type(), len(docs), len(docs)))

	for i, doc := range docs {
		elem := reflect.New(resultType)

		b, err1 := json.Marshal(doc.Fields)
		if err1 != nil {
			err = err1
			return
		}
		err1 = json.Unmarshal(b, elem.Interface())
		if err1 != nil {
			err = err1
			return
		}

		results.Index(i).Set(elem.Elem())
	}

	return
}

func (r *BaseRepository) List(c context.Context, query *entity.CursorQuery, ent entity.Entity, resultPtr interface{}) (extra *entity.CursorExtra, err error) {
	searchClient, err1 := r.Client(c)
	if err1 != nil {
		err = err1
		return
	}

	ms, err1 := breflect.GetStructInfo(ent, nil)
	if err1 != nil {
		err = err1
		return
	}

	typ := breflect.TheNamingStrategy.Table(ms.Name)
	index := r.DataSourceProvider.ProvideTable(c, typ)

	var docs []*search.Document
	docs, extra, err = searchClient.List(c, query, index, typ)
	if err != nil {
		return
	}

	results := reflect.ValueOf(resultPtr).Elem()
	resultType := results.Type().Elem()
	results.Set(reflect.MakeSlice(results.Type(), len(docs), len(docs)))

	for i, doc := range docs {
		elem := reflect.New(resultType)

		b, err := json.Marshal(doc.Fields)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(b, elem.Interface())
		if err != nil {
			return nil, err
		}

		results.Index(i).Set(elem.Elem())
	}

	// TODO
	return extra, err
}

func (r *BaseRepository) Connection(c context.Context, query *entity.ConnectionQuery, ent entity.Entity) (*entity.Connection, error) {
	searchClient, err := r.Client(c)
	if err != nil {
		return nil, err
	}

	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return nil, err
	}

	typ := breflect.TheNamingStrategy.Table(ms.Name)
	index := r.DataSourceProvider.ProvideTable(c, typ)

	conn, err := searchClient.Connection(c, query, index, typ)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
