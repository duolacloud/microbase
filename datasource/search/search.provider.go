package search

import (
	"github.com/duolacloud/microbase/proto/search"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/config"
)

func NewSearchService(client client.Client, config config.Config) (search.SearchService, error) {
	serviceName := config.Get("service", "search").String("com.microbase.srv.search")

	return search.NewSearchService(serviceName, client), nil
}
