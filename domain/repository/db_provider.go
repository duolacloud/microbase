package repository

import (
	"context"

	"github.com/duolacloud/microbase/multitenancy"
)

type DBProvider interface {
	Provide(ctx context.Context) (interface{}, error)
}

type MultitenancyProvider struct {
	tenancy multitenancy.Tenancy
}

func NewMultitenancyProvider(tenancy multitenancy.Tenancy) *MultitenancyProvider {
	return &MultitenancyProvider{
		tenancy,
	}
}

func (p *MultitenancyProvider) Provide(c context.Context) (interface{}, error) {
	tenantName, _ := multitenancy.FromContext(c)
	db, err := p.tenancy.ResourceFor(c, tenantName)
	return db, err
}
