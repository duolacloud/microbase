package memory_test

import (
	"log"
	"testing"
	"time"

	"github.com/duolacloud/microbase/cache"
	"github.com/duolacloud/microbase/cache/memory"

	"github.com/stretchr/testify/assert"
)

type User struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func TestBasic(t *testing.T) {
	s := memory.NewCache()
	s.Init()

	key := "hello"
	value := &User{Name: "李小龙", CreatedAt: time.Now()}

	err := s.Set(key, value, cache.WriteExpiry(time.Millisecond*1000))
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 500)

	var value1 *User
	ok := s.Get(key, &value1)
	log.Printf("get key: %v, value: %v, ok: %v", key, value1, ok)

	time.Sleep(time.Millisecond * 2000)

	var value2 *User
	ok = s.Get(key, &value2)
	log.Printf("get %s, result: %v", key, ok)
}
