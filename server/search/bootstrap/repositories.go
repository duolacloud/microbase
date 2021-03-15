package bootstrap

import (
	"github.com/duolacloud/microbase/service/search/repositories/elastic"
	"go.uber.org/fx"
)

var Repositories = fx.Provide(
	elastic.NewIndexRepository,
	elastic.NewDocumentRepository,
)

var RepositoryOpts = fx.Options(
	Repositories,
)
