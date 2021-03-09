package providers

import (
	"context"

	"github.com/duolacloud/microbase/logger"
	"github.com/urfave/cli/v2"

	_ "github.com/micro/go-plugins/registry/consul/v2"

	"github.com/micro/go-plugins/wrapper/monitoring/prometheus/v2"
	xopentracing "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	xxxmicro_opentracing "github.com/xxxmicro/base/opentracing"

	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"
	"github.com/micro/go-plugins/wrapper/validator/v2"
	"go.uber.org/fx"

	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2/server/grpc"
)

func NewMicroService(c *cli.Context) micro.Service {
	serviceName := c.String("server_name")
	serviceVersion := c.String("service_version")

	// use grpc server
	server := grpc.NewServer(server.WrapHandler(validator.NewHandlerWrapper()))
	QPS := 5000

	srv := micro.NewService(
		micro.Name(serviceName),
		micro.RegisterTTL(time.Minute),
		micro.RegisterInterval(time.Second*30),
		micro.WrapHandler(validator.NewHandlerWrapper()),
		micro.WrapHandler(xopentracing.NewHandlerWrapper(xxxmicro_opentracing.GlobalTracerWrapper())),
		micro.WrapHandler(prometheus.NewHandlerWrapper(prometheus.ServiceName(serviceName), prometheus.ServiceVersion(serviceVersion))),
		micro.WrapHandler(ratelimit.NewHandlerWrapper(QPS)),
		micro.WrapSubscriber(xopentracing.NewSubscriberWrapper(xxxmicro_opentracing.GlobalTracerWrapper())),
		micro.Server(server),
	)

	return srv
}

func StartMicroService(lifecycle fx.Lifecycle, srv micro.Service) {
	// TODO o := service.Options()
	// TODO micro.Broker(broker)(&o)

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Infof("service run")
			return srv.Run()
		},
	})
}

func MakeMicroServiceOpts(c *cli.Context) fx.Option {
	return fx.Options(
		fx.Provide(func() micro.Service {
			return NewMicroService(c)
		}),
		fx.Invoke(StartMicroService),
	)
}
