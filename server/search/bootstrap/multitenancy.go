package bootstrap

import (
	"github.com/duolacloud/microbase/multitenancy"
	"go.uber.org/fx"
)

var Multitenancy = fx.Provide(
	fx.Annotated{
		Name:   "elastic",
		Target: multitenancy.NewCachedTenancy,
	},
)

var MultitenancyOpts = fx.Options(
	Multitenancy,
)
