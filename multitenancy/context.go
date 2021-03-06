package multitenancy

import (
	"context"

	"github.com/micro/go-micro/v2/metadata"
)

const (
	TenantId = "tenant-id"
)

func FromContext(ctx context.Context) (string, bool) {
	/*
		tenantId, ok := ctx.Value(TenantId).(string)
		if ok {
			return tenantId, true
		}
	*/

	md, ok := metadata.FromContext(ctx)
	if ok {
		tenantId, ok := md.Get(TenantId)
		if ok {
			return tenantId, ok
		}
	}
	return "", false
}

func WithContext(ctx context.Context, tenantId string) context.Context {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(metadata.Metadata)
	}
	md.Set(TenantId, tenantId)

	ctx = metadata.NewContext(ctx, md)
	return ctx
}
