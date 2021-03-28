package cache

import (
	"github.com/duolacloud/microbase/cache/redis"
	"github.com/micro/go-micro/v2/config"
)

func NewCacheProvider(config config.Config) Cache {
	addrs := config.Get("cache", "addrs").StringSlice([]string{":6379"})

	var cache Cache
	switch driver {
	case "redis":
		cache := redis.NewCache(redis.WithAddrs(addrs...))
	default:
		cache := memory.NewCache()
	}

	return cache
}
