package elasticsearch

import (
	"context"
	"errors"
	"fmt"

	"github.com/duolacloud/microbase/database"
	"github.com/duolacloud/microbase/multitenancy"
	_reflect "github.com/duolacloud/microbase/reflect"
	"github.com/olivere/elastic/v6"
	// "github.com/olivere/elastic/v7"
)

type tenancy struct {
}

func NewElasticsearchTenancy(client *elastic.Client, entityMap database.EntityMap) multitenancy.Tenancy {
	var clientCreateFn = func(ctx context.Context, tenantId string) (multitenancy.Resource, error) {
		err := autoMigrate(ctx, client, entityMap, tenantId)
		if err != nil {
			return nil, err
		}

		return client, nil
	}

	var clientCloseFunc = func(resource multitenancy.Resource) {

	}

	return multitenancy.NewCachedTenancy(clientCreateFn, clientCloseFunc)
}

func indexName(entityName, tenantId string) string {
	if len(tenantId) == 0 {
		return entityName
	}

	return fmt.Sprintf("%s_%s", entityName, tenantId)
}

func autoMigrate(c context.Context, client *elastic.Client, entityMap database.EntityMap, tenantId string) error {
	for _, entity := range entityMap.GetEntities() {
		structInfo, err := _reflect.GetStructInfo(entity, nil)
		if err != nil {
			return err
		}

		indexName := indexName(_reflect.TheNamingStrategy.Table(structInfo.Name), tenantId)
		// Check if index exists
		indexExists, err := client.IndexExists(indexName).Do(c)
		if err != nil {
			return err
		}

		if !indexExists {
			r, err := client.CreateIndex(indexName).Do(c)
			if err != nil {
				return err
			}

			if !r.Acknowledged {
				return errors.New(fmt.Sprintf("expected IndicesCreateResult.Acknowledged true; got %v", r.Acknowledged))
			}
		}
	}

	return nil
}
