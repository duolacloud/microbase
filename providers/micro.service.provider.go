package providers

import (
	"context"

	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	_ "github.com/micro/go-plugins/registry/consul/v2"
	"github.com/micro/go-plugins/wrapper/monitoring/prometheus/v2"
	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"
	xopentracing "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	"github.com/micro/go-plugins/wrapper/validator/v2"
	"go.uber.org/fx"

	// "github.com/micro/go-plugins/wrapper/validator/v2"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	xxxmicro_opentracing "github.com/duolacloud/microbase/opentracing"
)

func NewMicroServiceFn(serviceName string, serviceVersion string) interface{} {
	return func() (micro.Service, *cli.Context) {
		return newMicroService(serviceName, serviceVersion)
	}
}

func newMicroService(serviceName string, serviceVersion string) (micro.Service, *cli.Context) {

	// use grpc server
	// server := grpc.NewServer(server.WrapHandler(validator.NewHandlerWrapper()))
	QPS := 5000

	service := micro.NewService(
		micro.Name(serviceName),
		micro.Version(serviceVersion),
		micro.RegisterTTL(time.Minute),
		micro.RegisterInterval(time.Second*30),
		micro.WrapHandler(validator.NewHandlerWrapper()),
		micro.WrapHandler(xopentracing.NewHandlerWrapper(xxxmicro_opentracing.GlobalTracerWrapper())),
		micro.WrapHandler(prometheus.NewHandlerWrapper(prometheus.ServiceName(serviceName), prometheus.ServiceVersion(serviceVersion))),
		micro.WrapHandler(ratelimit.NewHandlerWrapper(QPS)),
		micro.WrapSubscriber(xopentracing.NewSubscriberWrapper(xxxmicro_opentracing.GlobalTracerWrapper())),
		// micro.Server(server),
		micro.Flags(
			&cli.StringFlag{
				Name:  "apollo_namespace",
				Usage: "apollo_namespace",
			},
			&cli.StringFlag{
				Name:  "apollo_address",
				Usage: "apollo_address",
			},
			&cli.StringFlag{
				Name:  "apollo_app_id",
				Usage: "apollo_app_id",
			},
			&cli.StringFlag{
				Name:  "apollo_cluster",
				Usage: "apollo_cluster",
			},
			&cli.StringFlag{
				Name:    "prometheus_addr",
				Usage:   "prometheus_addr",
				EnvVars: []string{"PROMETHEUS_ADDR"},
				Value:   ":16627",
			},
		),
	)

	var cc *cli.Context
	service.Init(
		micro.Action(func(c *cli.Context) error {
			logger.Log(logger.DebugLevel, "service.Init")
			cc = c

			if len(c.String("prometheus_addr")) > 0 {
				prometheusBoot(c.String("prometheus_addr"))
			}

			return nil
		}),
	)

	return service, cc
}

func StartMicroService(lifecycle fx.Lifecycle, service micro.Service /* broker broker.Broker,*/, tracer opentracing.Tracer) {
	// TODO o := service.Options()
	// TODO micro.Broker(broker)(&o)

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Log(logger.DebugLevel, "lifecycle.OnStart")
			return service.Run()
		},
	})
}

func prometheusBoot(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			logger.Fatal("ListenAndServe: ", err)
		}
	}()
}

func MakeMicroServiceOpts(serviceName string, serviceVersion string) fx.Option {
	return fx.Options(
		fx.Provide(NewMicroServiceFn(serviceName, serviceVersion)),
		fx.Invoke(StartMicroService),
	)
}
