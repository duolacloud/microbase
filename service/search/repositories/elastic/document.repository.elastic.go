package elastic

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/duolacloud/microbase/client/search"
	"github.com/duolacloud/microbase/domain/entity"
	"github.com/duolacloud/microbase/domain/repository"
	"github.com/duolacloud/microbase/domain/repository/elasticsearch"
	"github.com/duolacloud/microbase/logger"
	"github.com/duolacloud/microbase/service/search/repositories"

	"github.com/olivere/elastic/v6"
	// "github.com/olivere/elastic/v7"
)

type DocumentRepository struct {
	DataSourceProvider repository.DataSourceProvider
}

func NewDocumentRepository(dataSourceProvider repository.DataSourceProvider) repositories.DocumentRepository {
	return &DocumentRepository{
		dataSourceProvider,
	}
}

func (r *DocumentRepository) client(c context.Context) (*elastic.Client, error) {
	client, err := r.DataSourceProvider.ProvideDB(c)
	if err != nil {
		return nil, err
	}
	return client.(*elastic.Client), nil
}

func (r *DocumentRepository) Create(c context.Context, doc *search.Document) error {
	client, err := r.client(c)
	if err != nil {
		return err
	}

	index := r.DataSourceProvider.ProvideTable(c, doc.Index)

	_, err = client.Update().
		Index(index).
		Type(doc.Type).
		Id(doc.Fields["id"].(string)).
		Doc(doc.Fields).
		DocAsUpsert(true).
		Do(c)
	return err
}

func (r *DocumentRepository) Upsert(c context.Context, doc *search.Document) error {
	client, err := r.client(c)
	if err != nil {
		return err
	}
	index := r.DataSourceProvider.ProvideTable(c, doc.Index)

	_, err = client.Update().
		Index(index).
		Type(doc.Type).
		Id(doc.Fields["id"].(string)).
		DocAsUpsert(true).
		Doc(doc.Fields).
		Do(c)

	return err
}

func (r *DocumentRepository) Update(c context.Context, doc *search.Document) error {
	b, _ := json.Marshal(doc.Fields)
	logger.Infof("DocumentRepository Update: %v", string(b))

	client, err := r.client(c)
	if err != nil {
		return err
	}
	index := r.DataSourceProvider.ProvideTable(c, doc.Index)

	id, ok := doc.Fields["id"].(string)
	if !ok {
		return errors.New("update need id")
	}

	_, err = client.Update().
		Index(index).
		Type(doc.Type).
		Id(id).
		Doc(doc.Fields).
		Do(c)
	return err
}

func (r *DocumentRepository) Get(c context.Context, index, typ, id string) (*search.Document, error) {
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

	doc := &search.Document{
		Index: index,
		Type:  typ,
	}
	err = json.Unmarshal(*res.Source, &doc.Fields)
	log.Printf("DocumentRepository Get, index: %s, type: %s, id: %s, res: %v", index, typ, id, doc.Fields)

	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *DocumentRepository) Delete(c context.Context, index, typ, id string) error {
	client, err := r.client(c)
	if err != nil {
		return err
	}

	log.Printf("DocumentRepository.Delete, index: %s, type: %s, id: %s", index, typ, id)
	_, err = client.Delete().
		Index(index).
		Type(typ).
		Id(id).
		Do(c)

	if elastic.IsNotFound(err) {
		return nil
	}

	return err
}

func (r *DocumentRepository) Page(c context.Context, query *entity.PageQuery, index, typ string) (docs []*search.Document, total int64, err error) {
	client, err1 := r.client(c)
	if err1 != nil {
		err = err1
		return
	}

	paginator := elasticsearch.NewPaginator(client)
	docs, total, err = paginator.Paginate(c, query, index, typ)
	return
}

func (r *DocumentRepository) List(c context.Context, query *entity.CursorQuery, index, typ string) (docs []*search.Document, extra *entity.CursorExtra, err error) {
	client, err1 := r.client(c)
	if err1 != nil {
		err = err1
		return
	}

	paginator := elasticsearch.NewCursorPaginator(client)

	docs, extra, err = paginator.Paginate(c, query, index, typ)
	return
}

func (r *DocumentRepository) Connection(c context.Context, query *entity.ConnectionQuery, index, typ string) (*entity.Connection, error) {
	client, err := r.client(c)
	if err != nil {
		return nil, err
	}

	paginator := elasticsearch.NewConnectionPaginator(client)

	return paginator.Paginate(c, query, index, typ)
}
