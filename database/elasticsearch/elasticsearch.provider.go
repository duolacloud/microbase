package elasticsearch

import (
	// elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/micro/go-micro/v2/config"
)

func NewElasticSearchClient(config config.Config) (*elasticsearch.Client, error) {
	addrs := config.Get("addrs").StringSlice([]string{
		"http://localhost:9200",
		"http://localhost:9201",
	})

	cfg := elasticsearch.Config{
		Addresses: addrs,
		// ...
	}

	es, err := elasticsearch.NewClient(cfg)
	return es, err
}
