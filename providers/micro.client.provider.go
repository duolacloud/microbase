package providers

import (
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/util/wrapper"
	"github.com/micro/go-plugins/registry/consul/v2"
	xopentracing "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/urfave/cli/v2"
)

func NewMicroClient(cli *cli.Context, tracer opentracing.Tracer) client.Client {
	registryAddress := cli.StringSlice("registry_address")

	reg := consul.NewRegistry(func(op *registry.Options) {
		op.Addrs = registryAddress
	})

	c := grpc.NewClient(
		client.Registry(reg),
		client.Selector(selector.DefaultSelector),
		client.Wrap(xopentracing.NewClientWrapper(tracer)),
		// client.Wrap(hystrix.NewClientWrapper()),
	)

	cacheFn := func() *client.Cache { return c.Options().Cache }
	c = wrapper.CacheClient(cacheFn, c)

	return c
}
