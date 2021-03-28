package redis_test

import (
	"testing"
	"time"

	"github.com/duolacloud/microbase/cache"
	"github.com/duolacloud/microbase/cache/redis"
	"github.com/stretchr/testify/assert"
)

type User struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func TestBasic(t *testing.T) {
	s := redis.NewCache(redis.WithAddrs(":6379"))
	s.Init()

	key := "hello"
	value := &User{Name: "李小龙", CreatedAt: time.Now()}

	err := s.Set(key, value, cache.WriteExpiry(time.Millisecond*1000))
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 500)

	var value1 *User
	ok := s.Get(key, &value1)
	assert.Equal(t, true, ok, "Expected no records in redis store")

	time.Sleep(time.Millisecond * 2000)

	var value2 *User
	ok = s.Get(key, &value2)
	assert.Equal(t, false, ok, "Expected no records in redis store")
}
