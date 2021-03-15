package elasticsearch

import (
	"context"
	"encoding/json"

	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"

	"github.com/olivere/elastic/v6"
	// "github.com/olivere/elastic/v7"
)

type DocumentRepository struct {
	DataSourceProvider repository.DataSourceProvider
}

func NewDocumentRepository(dataSourceProvider repository.DataSourceProvider) repository.DocumentRepository {
	return &DocumentRepository{dataSourceProvider}
}

func (r *DocumentRepository) client(c context.Context) (*elastic.Client, error) {
	client, err := r.DataSourceProvider.ProvideDB(c)
	if err != nil {
		return nil, err
	}
	return client.(*elastic.Client), nil
}

func (r *DocumentRepository) Create(c context.Context, doc *entity.Document) error {
	client, err := r.client(c)
	if err != nil {
		return err
	}

	index := r.DataSourceProvider.ProvideTable(c, doc.Index)

	_, err = client.Update().
		Index(index).
		Type(doc.Type).
		Id(doc.ID).
		Upsert(doc.Fields).
		Do(c)
	return err
}

func (r *DocumentRepository) Upsert(c context.Context, doc *entity.Document) error {
	client, err := r.client(c)
	if err != nil {
		return err
	}
	index := r.DataSourceProvider.ProvideTable(c, doc.Index)

	_, err = client.Update().
		Index(index).
		Type(doc.Type).
		Id(doc.ID).
		DocAsUpsert(true).
		Doc(doc.Fields).
		Do(c)

	return err
}

func (r *DocumentRepository) Update(c context.Context, doc *entity.Document) error {
	client, err := r.client(c)
	if err != nil {
		return err
	}
	index := r.DataSourceProvider.ProvideTable(c, doc.Index)

	_, err = client.Update().
		Index(index).
		Type(doc.Type).
		Id(doc.ID).
		Doc(doc.Fields).
		Do(c)
	return err
}

func (r *DocumentRepository) Get(c context.Context, id string, index, typ string) (*entity.Document, error) {
	client, err := r.client(c)
	if err != nil {
		return nil, err
	}
	index = r.DataSourceProvider.ProvideTable(c, index)

	res, err := client.Get().
		Index(index).
		Type(typ).
		Id(id).
		Do(c)
	if err != nil {
		return nil, err
	}

	if !res.Found {
		return nil, nil
	}

	doc := &entity.Document{
		ID:    id,
		Index: index,
		Type:  typ,
	}
	err = json.Unmarshal(*res.Source, &doc.Fields)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *DocumentRepository) Delete(c context.Context, id string, index, typ string) error {
	client, err := r.client(c)
	if err != nil {
		return err
	}
	index = r.DataSourceProvider.ProvideTable(c, index)

	_, err = client.Delete().
		Index(index).
		Type(typ).
		Id(id).
		Do(c)
	return err
}

func (r *DocumentRepository) Page(c context.Context, query *entity.PageQuery, index, typ string) (docs []*entity.Document, total int, pageCount int, err error) {
	paginator := NewPaginator(r.DataSourceProvider)
	docs, total, pageCount, err = paginator.Paginate(c, query, index, typ)
	return
}

func (r *DocumentRepository) List(c context.Context, query *entity.CursorQuery, index, typ string) (docs []*entity.Document, extra *entity.CursorExtra, err error) {
	client, err1 := r.client(c)
	if err1 != nil {
		err = err1
		return
	}
	index = r.DataSourceProvider.ProvideTable(c, index)

	paginator := NewCursorPaginator(client)

	docs, extra, err = paginator.Paginate(c, query, index, typ)
	return
}

func (r *DocumentRepository) Connection(c context.Context, query *entity.ConnectionQuery, index, typ string) (*entity.Connection, error) {
	paginator := NewConnectionPaginator(r.DataSourceProvider)

	return paginator.Paginate(c, query, index, typ)
}
