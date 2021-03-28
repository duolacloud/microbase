package providers

import (
	"github.com/duolacloud/microbase/cache"
	"github.com/duolacloud/microbase/cache/memory"
	"github.com/duolacloud/microbase/cache/redis"
	"github.com/micro/go-micro/v2/config"
)

func NewCacheProvider(config config.Config) cache.Cache {
	driver := config.Get("cache", "addrs").String("redis")
	prefix := config.Get("cache", "prefix").String("")

	var c cache.Cache
	switch driver {
	case "redis":
		addrs := config.Get("cache", "addrs").StringSlice([]string{":6379"})
		c = redis.NewCache(cache.WithPrefix(prefix), redis.WithAddrs(addrs...))
	default:
		c = memory.NewCache(cache.WithPrefix(prefix))
	}

	return c
}
