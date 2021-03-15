module github.com/duolacloud/microbase

go 1.14

require (
	contrib.go.opencensus.io/integrations/ocsql v0.1.7 // indirect
	entgo.io/ent v0.6.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-git/go-git/v5 v5.1.0 // indirect
	github.com/go-sql-driver/mysql v1.5.1-0.20200311113236-681ffa848bae
	github.com/golang/protobuf v1.4.3
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/inflection v1.0.0
	github.com/micro/cli/v2 v2.1.2 // indirect
	github.com/micro/go-micro v1.18.0 // indirect
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-plugins v1.5.1 // indirect
	github.com/micro/go-plugins/logger/zap/v2 v2.9.1
	github.com/micro/go-plugins/registry/consul/v2 v2.9.1
	github.com/micro/go-plugins/wrapper/monitoring/prometheus/v2 v2.9.1
	github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2 v2.9.1
	github.com/micro/go-plugins/wrapper/trace/opentracing/v2 v2.9.1
	github.com/micro/go-plugins/wrapper/validator/v2 v2.9.1
	github.com/micro/micro v1.18.0 // indirect
	github.com/micro/micro/v2 v2.9.3 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/mwitkow/go-proto-validators v0.3.2 // indirect
	github.com/olivere/elastic v6.2.35+incompatible
	github.com/olivere/elastic/v6 v6.2.1
	github.com/olivere/elastic/v7 v7.0.22
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/common v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/stretchr/testify v1.7.0
	github.com/thoas/go-funk v0.7.0
	github.com/transaction-wg/seata-golang v0.1.9
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/urfave/cli/v2 v2.2.0
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/vmihailenco/msgpack/v5 v5.2.0
	github.com/xxxmicro/base v0.1.35
	github.com/xxxmicro/go-micro-apollo-plugin v1.1.4
	go.uber.org/fx v1.13.1
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	google.golang.org/appengine v1.6.1
	google.golang.org/grpc v1.33.2
	google.golang.org/protobuf v1.25.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
)

replace (
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)