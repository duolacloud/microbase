package providers

import (
	"github.com/duolacloud/microbase/config"
	xconfig "github.com/duolacloud/microbase/config"
	xsource "github.com/duolacloud/microbase/config/source"
	"github.com/duolacloud/microbase/opentracing/jaeger"
	"go.uber.org/fx"
)

var Framework = fx.Provide(
	config.NewAppConfig,
	xsource.NewSourceProvider,
	xconfig.NewConfigProvider,
	jaeger.NewTracerProvider,
)

var FrameworkOpts = fx.Options(
	Framework,
	fx.Invoke(InitLogger),
	fx.Invoke(StartPrometheus),
)
