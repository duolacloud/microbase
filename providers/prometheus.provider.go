package providers

import (
	"context"
	"net/http"

	"github.com/duolacloud/microbase/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"

	"go.uber.org/fx"
)

func prometheusBoot(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			logger.Fatal("ListenAndServe: ", err)
		}
	}()
}

func StartPrometheus(lifecycle fx.Lifecycle, cli *cli.Context, tracer opentracing.Tracer) {
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			paddr := cli.String("prometheus_addr")
			if len(paddr) > 0 {
				logger.Debugf("prometheus run")

				prometheusBoot(paddr)
			}
			return nil
		},
	})
}
