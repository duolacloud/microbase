package elasticsearch

import (
	// elasticsearch "github.com/elastic/go-elasticsearch/v6"
	"github.com/micro/go-micro/v2/config"
	"github.com/olivere/elastic/v6"
)

func NewElasticSearchClient(config config.Config) (*elastic.Client, error) {
	addrs := config.Get("addrs").StringSlice([]string{
		"http://localhost:9200",
		"http://localhost:9201",
	})

	client, err := elastic.NewClient(
		elastic.SetURL(addrs...),
		elastic.SetSniff(false), // 如果打开 sniff 后，会嗅探到 docker 内网地址，导致无法连接节点
	)

	return client, err
}
