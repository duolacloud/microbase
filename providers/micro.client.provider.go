package providers

import (
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/util/wrapper"
	xopentracing "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"
	opentracing "github.com/opentracing/opentracing-go"
)

func NewMicroClient(tracer opentracing.Tracer) client.Client {
	c := grpc.NewClient(
		client.Selector(selector.DefaultSelector),
		client.Wrap(xopentracing.NewClientWrapper(tracer)),
		// client.Wrap(hystrix.NewClientWrapper()),
	)

	cacheFn := func() *client.Cache { return c.Options().Cache }
	c = wrapper.CacheClient(cacheFn, c)

	return c
}
