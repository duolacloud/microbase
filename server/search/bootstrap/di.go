package bootstrap

import (
	framework "github.com/duolacloud/microbase/providers"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

func Setup(c *cli.Context) *fx.App {
	return fx.New(
		fx.Provide(
			func() *cli.Context {
				return c
			},
		),
		framework.FrameworkOpts,
		HandlerOpts,
		MultitenancyOpts,
		framework.MakeMicroServiceOpts(c))
}
