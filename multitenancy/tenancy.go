package multitenancy

import (
	"context"
	"runtime"
	"sync"
)

type Resource interface{}

type Tenancy interface {
	ResourceFor(ctx context.Context, tenantName string) (Resource, error)
}

type cachedTenancy struct {
	resourceCreateFunc func(ctx context.Context, tenantName string) (Resource, error)
	resourceCloseFunc  func(resource Resource)
	resources          map[string]Resource
	mu                 sync.RWMutex
}

func NewCachedTenancy(
	resourceCreateFunc func(ctx context.Context, tenantName string) (Resource, error),
	resourceCloseFunc func(resource Resource),
) Tenancy {
	tenancy := &cachedTenancy{
		resourceCreateFunc: resourceCreateFunc,
		resourceCloseFunc:  resourceCloseFunc,
		resources:          map[string]Resource{},
	}

	runtime.SetFinalizer(tenancy, func(tenancy *cachedTenancy) {
		for _, resource := range tenancy.resources {
			resourceCloseFunc(resource)
		}
	})
	return tenancy
}

func (c *cachedTenancy) ResourceFor(ctx context.Context, tenantName string) (Resource, error) {
	c.mu.RLock()
	resource, ok := c.resources[tenantName]
	if ok {
		return resource, nil
	}
	c.mu.RUnlock()

	if ok {
		return resource, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if resource, ok := c.resources[tenantName]; ok {
		return resource, nil
	}

	resource, err := c.resourceCreateFunc(ctx, tenantName)
	if err != nil {
		return resource, err
	}

	c.resources[tenantName] = resource
	return resource, nil
}
