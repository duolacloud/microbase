package bootstrap

import (
	"github.com/duolacloud/microbase/service/search/handlers"
	"github.com/duolacloud/microbase/service/search/providers"
	"go.uber.org/fx"
)

var Handlers = fx.Provide(
	handlers.NewSearchHandler,
)

var HandlerOpts = fx.Options(
	Handlers,
	fx.Invoke(providers.RegisterHandlers),
)
