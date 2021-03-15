package elasticsearch

import (
	"context"

	"github.com/duolacloud/microbase/database"
	"github.com/duolacloud/microbase/multitenancy"
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

func autoMigrate(c context.Context, client *elastic.Client, entityMap database.EntityMap, tenantId string) error {
	for _, entity := range entityMap.GetEntities() {
		indexModel := NewIndexModel(client, entity, tenantId)
		if err := indexModel.CreateIndex(c); err != nil {
			return err
		}
	}

	return nil
}
