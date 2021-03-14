package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"
	breflect "github.com/duolacloud/microbase/reflect"
	"github.com/mitchellh/mapstructure"
	"github.com/olivere/elastic"
	"github.com/prometheus/common/log"
	"github.com/thoas/go-funk"
)

type BaseRepository struct {
	DataSourceProvider repository.DataSourceProvider
	DocumentRepository repository.DocumentRepository
}

func NewBaseRepository(dataSourceProvider repository.DataSourceProvider) repository.BaseRepository {
	return &BaseRepository{
		DataSourceProvider: dataSourceProvider,
		DocumentRepository: NewDocumentRepository(dataSourceProvider),
	}
}

func (r *BaseRepository) Client(c context.Context) (*elastic.Client, error) {
	client, err := r.DataSourceProvider.ProvideDB(c)
	if err != nil {
		return nil, err
	}
	return client.(*elastic.Client), nil
}

func (r *BaseRepository) Create(c context.Context, ent entity.Entity) error {
	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return err
	}
	index := breflect.TheNamingStrategy.Table(ms.Name)
	typ := index

	// 命名约定，必须有 id 字段
	id, ok := funk.Get(ent, "id").(string)
	if !ok {
		return errors.New(fmt.Sprintf("no id field for entity %v", ent))
	}

	var fields map[string]interface{}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &fields,
		TagName: "json",
	})
	if err != nil {
		return err
	}

	err = decoder.Decode(ent)
	if err != nil {
		return err
	}

	return r.DocumentRepository.Create(c, &entity.Document{
		ID:     id,
		Index:  index,
		Type:   typ,
		Fields: fields,
	})
}

func (r *BaseRepository) Upsert(c context.Context, ent entity.Entity) (*repository.ChangeInfo, error) {
	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return nil, err
	}
	index := breflect.TheNamingStrategy.Table(ms.Name)
	typ := index

	// 命名约定，必须有 id 字段
	id, ok := funk.Get(ent, "ID").(string)
	if !ok {
		return nil, errors.New(fmt.Sprintf("no id field for entity %v", ent))
	}

	var fields map[string]interface{}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &fields,
		TagName: "json",
	})
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(ent)
	if err != nil {
		return nil, err
	}

	err = r.DocumentRepository.Upsert(c, &entity.Document{
		ID:     id,
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
	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return err
	}
	index := ms.Name
	typ := ms.Name

	// 命名约定，必须有 id 字段
	id, ok := funk.Get(ent, "id").(string)
	if !ok {
		return errors.New(fmt.Sprintf("no id field for entity %v", ent))
	}

	var fields map[string]interface{}
	err = mapstructure.Decode(data, fields)
	if err != nil {
		return err
	}

	return r.DocumentRepository.Update(c, &entity.Document{
		ID:     id,
		Index:  index,
		Type:   typ,
		Fields: fields,
	})
}

func (r *BaseRepository) Get(c context.Context, ent entity.Entity) error {
	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return err
	}
	index := ms.Name
	typ := ms.Name

	// 命名约定，必须有 id 字段
	id, ok := funk.Get(ent, "id").(string)
	if !ok {
		return errors.New(fmt.Sprintf("no id field for entity %v", ent))
	}

	doc, err := r.DocumentRepository.Get(c, id, index, typ)
	if err != nil {
		return err
	}

	if doc == nil {
		return nil
	}

	err = mapstructure.Decode(doc.Fields, ent)
	if err != nil {
		return err
	}
	return nil
}

func (r *BaseRepository) Delete(c context.Context, ent entity.Entity) error {
	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return err
	}
	index := ms.Name
	typ := ms.Name

	// 命名约定，必须有 id 字段
	id, ok := funk.Get(ent, "id").(string)
	if !ok {
		return errors.New(fmt.Sprintf("no id field for entity %v", ent))
	}

	return r.DocumentRepository.Delete(c, id, index, typ)
}

func (r *BaseRepository) Page(c context.Context, ent entity.Entity, query *entity.PageQuery, resultPtr interface{}) (total int, pageCount int, err error) {
	ms, err1 := breflect.GetStructInfo(ent, nil)
	if err1 != nil {
		err = err1
		return
	}

	index := ms.Name
	typ := ms.Name

	paginator := NewPaginator(r.DataSourceProvider)
	var docs []*entity.Document
	docs, total, pageCount, err = paginator.Paginate(c, query, index, typ)
	if err != nil {
		return
	}

	log.Info(docs)

	return
}

func (r *BaseRepository) List(c context.Context, query *entity.CursorQuery, ent entity.Entity, resultPtr interface{}) (extra *entity.CursorExtra, err error) {
	ms, err1 := breflect.GetStructInfo(ent, nil)
	if err1 != nil {
		err = err1
		return
	}

	index := breflect.TheNamingStrategy.Table(ms.Name)
	typ := index

	var docs []*entity.Document
	docs, extra, err = r.DocumentRepository.List(c, query, index, typ)
	if err != nil {
		return
	}

	results := reflect.ValueOf(resultPtr).Elem()
	resultType := results.Type().Elem()
	results.Set(reflect.MakeSlice(results.Type(), len(docs), len(docs)))

	for i, doc := range docs {
		elem := reflect.New(resultType)

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:  elem.Interface(),
			TagName: "json",
		})
		if err != nil {
			return nil, err
		}

		err = decoder.Decode(doc.Fields)
		if err != nil {
			return nil, err
		}

		results.Index(i).Set(elem.Elem())
	}

	// TODO
	return extra, err
}

func (r *BaseRepository) Connection(c context.Context, query *entity.ConnectionQuery, ent entity.Entity) (*entity.Connection, error) {
	ms, err := breflect.GetStructInfo(ent, nil)
	if err != nil {
		return nil, err
	}

	index := ms.Name
	typ := ms.Name

	paginator := NewConnectionPaginator(r.DataSourceProvider)

	return paginator.Paginate(c, query, index, typ)
}
