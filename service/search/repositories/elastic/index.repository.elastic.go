package elastic

import (
	"context"
	"errors"
	"log"

	"github.com/duolacloud/microbase/client/search"
	"github.com/duolacloud/microbase/domain/repository"
	"github.com/duolacloud/microbase/service/search/repositories"
	"github.com/olivere/elastic/v6"
)

type indexRepository struct {
	dataSourceProvider repository.DataSourceProvider
}

func (r *indexRepository) Client(c context.Context) (*elastic.Client, error) {
	client, err := r.dataSourceProvider.ProvideDB(c)
	if err != nil {
		return nil, err
	}
	return client.(*elastic.Client), nil
}

func NewIndexRepository(dataSourceProvider repository.DataSourceProvider) repositories.IndexRepository {
	return &indexRepository{
		dataSourceProvider,
	}
}

func (r *indexRepository) Create(c context.Context, index *search.Index) error {
	client, err := r.Client(c)
	if err != nil {
		return err
	}

	log.Printf("create index: %v", index.Mapping)
	_, err = client.CreateIndex(index.Name).BodyJson(index.Mapping).Do(c)
	if err != nil {
		return err
	}

	return nil
}

func (r *indexRepository) Delete(c context.Context, index string) error {
	client, err := r.Client(c)
	if err != nil {
		return err
	}

	rsp, err := client.DeleteIndex(index).Do(c)
	if err != nil {
		return err
	}

	if rsp.Acknowledged {
		return errors.New("Delete Index Acknowledged false")
	}

	return nil
}

func (r *indexRepository) IndexExists(c context.Context, index string) (bool, error) {
	client, err := r.Client(c)
	if err != nil {
		return false, err
	}

	return client.IndexExists(index).Do(c)
}
