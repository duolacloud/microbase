module github.com/xxxmicro/microbase

go 1.14

require (
	contrib.go.opencensus.io/integrations/ocsql v0.1.7
	github.com/elastic/go-elasticsearch/v7 v7.11.0
	github.com/go-git/go-git/v5 v5.1.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.0
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/inflection v1.0.0
	github.com/micro/cli/v2 v2.1.2
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-plugins/logger/zap/v2 v2.9.1
	github.com/micro/go-plugins/registry/consul/v2 v2.9.1
	github.com/micro/go-plugins/wrapper/monitoring/prometheus/v2 v2.9.1
	github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2 v2.9.1
	github.com/micro/go-plugins/wrapper/trace/opentracing/v2 v2.9.1
	github.com/micro/go-plugins/wrapper/validator/v2 v2.9.1
	github.com/mwitkow/go-proto-validators v0.3.2
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/satori/go.uuid v1.2.0
	github.com/stretchr/testify v1.6.1
	github.com/transaction-wg/seata-golang v0.1.9
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/xxxmicro/base v0.1.35
	github.com/xxxmicro/go-micro-apollo-plugin v1.1.4
	go.uber.org/fx v1.13.1
	go.uber.org/zap v1.15.0
	google.golang.org/protobuf v1.22.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
)
