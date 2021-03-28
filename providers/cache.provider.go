package providers

import (
	"github.com/duolacloud/microbase/cache/memory"
	"github.com/duolacloud/microbase/cache/redis"
	"github.com/micro/go-micro/v2/config"
)

func NewCacheProvider(config config.Config) Cache {
	addrs := config.Get("cache", "addrs").StringSlice([]string{":6379"})
	prefix := config.Get("cache", "prefix").StringSlice([]string{""})

	var cache Cache
	switch driver {
	case "redis":
		cache := redis.NewCache(cache.WithPrefix(prefix), redis.WithAddrs(addrs...))
	default:
		cache := memory.NewCache(cache.WithPrefix(prefix))
	}

	return cache
}
