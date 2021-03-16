package providers

import (
	client_search "github.com/duolacloud/microbase/client/search"
	datasource_search "github.com/duolacloud/microbase/datasource/search"
	"go.uber.org/fx"
)

var CloudNative = fx.Provide(
	datasource_search.NewSearchService,
	client_search.NewSearchClient,
)

var CloudNativeOpts = fx.Options(
	CloudNative,
)
