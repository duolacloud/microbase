package providers

import (
	"context"
	"log"

	"github.com/duolacloud/microbase/logger"
	"github.com/urfave/cli/v2"

	"github.com/micro/go-plugins/registry/consul/v2"

	xxxmicro_opentracing "github.com/duolacloud/microbase/opentracing"
	"github.com/micro/go-plugins/wrapper/monitoring/prometheus/v2"
	xopentracing "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"

	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"
	"github.com/micro/go-plugins/wrapper/validator/v2"
	"go.uber.org/fx"

	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
)

func NewMicroService(c *cli.Context) micro.Service {
	serviceName := c.String("service_name")
	serviceVersion := c.String("service_version")
	registryAddress := c.StringSlice("registry_address")
	logger.Infof("registryAddress: %s", registryAddress)
	// use grpc server
	// server := grpc.NewServer(server.WrapHandler(validator.NewHandlerWrapper()))
	QPS := 5000

	reg := consul.NewRegistry(func(op *registry.Options) {
		op.Addrs = registryAddress
	})

	srv := micro.NewService(
		micro.Name(serviceName),
		micro.Registry(reg),
		micro.RegisterTTL(time.Minute),
		micro.RegisterInterval(time.Second*30),
		micro.WrapHandler(validator.NewHandlerWrapper()),
		micro.WrapHandler(xopentracing.NewHandlerWrapper(xxxmicro_opentracing.GlobalTracerWrapper())),
		micro.WrapHandler(prometheus.NewHandlerWrapper(prometheus.ServiceName(serviceName), prometheus.ServiceVersion(serviceVersion))),
		micro.WrapHandler(ratelimit.NewHandlerWrapper(QPS)),
		micro.WrapSubscriber(xopentracing.NewSubscriberWrapper(xxxmicro_opentracing.GlobalTracerWrapper())),
		// micro.Server(server),
	)

	return srv
}

func StartMicroService(lifecycle fx.Lifecycle, srv micro.Service) {
	// TODO o := service.Options()
	// TODO micro.Broker(broker)(&o)

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Printf("service run")
			return srv.Run()
		},
	})
}

func MakeMicroServiceOpts(c *cli.Context) fx.Option {
	return fx.Options(
		fx.Provide(NewMicroService),
		fx.Invoke(StartMicroService),
	)
}
