package bootstrap

import (
	"github.com/duolacloud/microbase/datasource"
	"github.com/duolacloud/microbase/datasource/elasticsearch"
	"github.com/duolacloud/microbase/domain/repository"
	"go.uber.org/fx"
)

type EntityMap struct {
}

func (m *EntityMap) GetEntities() []interface{} {
	return []interface{}{}
}

func NewEntityMap() datasource.EntityMap {
	return &EntityMap{}
}

var Datasources = fx.Provide(
	elasticsearch.NewElasticSearchClient,
	elasticsearch.NewElasticSearchTenancy,
	repository.NewMultitenancyProvider,
	NewEntityMap,
)

var DatasourceOpts = fx.Options(
	Datasources,
)
