package bootstrap

import (
	"github.com/duolacloud/microbase/datasource/elasticsearch"
	"github.com/duolacloud/microbase/domain/repository"
	"github.com/duolacloud/microbase/service/search/providers"
	"go.uber.org/fx"
)

var Datasources = fx.Provide(
	elasticsearch.NewElasticSearchClient,
	elasticsearch.NewElasticSearchTenancy,
	repository.NewMultitenancyProvider,
	providers.NewEntityMap,
)

var DatasourceOpts = fx.Options(
	Datasources,
)
