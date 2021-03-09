package providers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/web"
	_ "github.com/micro/go-plugins/registry/consul/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

func NewMicroWebService(c *cli.Context) web.Service {
	return web.NewService(
		web.RegisterTTL(time.Minute),
		web.RegisterInterval(time.Second*30),
	)
}

func StartMicroWebService(lifecycle fx.Lifecycle, service web.Service, gin *gin.Engine) {
	service.Handle("/", gin)

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return service.Run()
		},
	})
}

func MakeMicroWebServiceOpts(c *cli.Context) fx.Option {
	return fx.Options(
		fx.Provide(func() web.Service {
			return NewMicroWebService(c)
		}),
		fx.Provide(NewMicroClient),
		fx.Invoke(StartMicroWebService),
	)
}
