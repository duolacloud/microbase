package repository

import (
	"context"
	"fmt"

	"github.com/duolacloud/microbase/multitenancy"
)

type DataSourceProvider interface {
	ProvideDB(ctx context.Context) (interface{}, error)
	ProvideTable(ctx context.Context, tableName string) string
}

type MultitenancyProvider struct {
	tenancy multitenancy.Tenancy
}

func NewMultitenancyProvider(tenancy multitenancy.Tenancy) *MultitenancyProvider {
	return &MultitenancyProvider{
		tenancy,
	}
}

func (p *MultitenancyProvider) ProvideDB(c context.Context) (interface{}, error) {
	tenantId, _ := multitenancy.FromContext(c)
	db, err := p.tenancy.ResourceFor(c, tenantId)
	return db, err
}

func (p *MultitenancyProvider) ProvideTable(c context.Context, tableName string) string {
	tenantId, _ := multitenancy.FromContext(c)

	if len(tenantId) == 0 {
		return tableName
	}

	return fmt.Sprintf("%s_%s", tableName, tenantId)
}
