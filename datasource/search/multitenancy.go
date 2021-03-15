package search

import (
	"context"

	"github.com/duolacloud/microbase/client/search"
	"github.com/duolacloud/microbase/datasource"
	"github.com/duolacloud/microbase/multitenancy"
)

type tenancy struct {
}

func NewSearchTenancy(searchClient search.SearchClient, entityMap datasource.EntityMap) multitenancy.Tenancy {
	var tenancyCreateFn = func(ctx context.Context, tenantId string) (multitenancy.Resource, error) {
		err := autoMigrate(ctx, searchClient, entityMap, tenantId)
		if err != nil {
			return nil, err
		}

		return searchClient, nil
	}

	var tenancyCloseFunc = func(resource multitenancy.Resource) {

	}

	return multitenancy.NewCachedTenancy(tenancyCreateFn, tenancyCloseFunc)
}

func autoMigrate(c context.Context, searchClient search.SearchClient, entityMap datasource.EntityMap, tenantId string) error {
	for _, entity := range entityMap.GetEntities() {
		indexModel := NewIndexModel(searchClient, entity, tenantId)
		if err := indexModel.CreateIndex(c); err != nil {
			return err
		}
	}

	return nil
}
