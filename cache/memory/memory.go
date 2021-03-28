package memory

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/duolacloud/microbase/cache"
	_cache "github.com/patrickmn/go-cache"
)

type MemoryCache struct {
	options cache.Options
	cache   *_cache.Cache
}

func NewCache(opts ...cache.Option) cache.Cache {
	options := cache.Options{
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	return &MemoryCache{
		options: options,
		cache:   _cache.New(24*time.Hour, 30*time.Second),
	}
}

func (m *MemoryCache) Init(opts ...cache.Option) error {
	for _, o := range opts {
		o(&m.options)
	}

	return nil
}

func (m *MemoryCache) Options() cache.Options {
	return m.options
}

func (m *MemoryCache) prefix(key string) string {
	if m.options.Prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", m.options.Prefix, key)
}

func (m *MemoryCache) String() string {
	return "memory"
}

func (m *MemoryCache) Get(key string, resultPtr interface{}, opts ...cache.ReadOption) bool {
	readOpts := cache.ReadOptions{}
	for _, o := range opts {
		o(&readOpts)
	}

	key = m.prefix(key)

	data, ok := m.cache.Get(key)
	if !ok {
		return false
	}

	v := reflect.ValueOf(resultPtr)
	v.Elem().Set(reflect.ValueOf(data))

	return true
}

func (m *MemoryCache) Set(key string, value interface{}, opts ...cache.WriteOption) error {
	writeOpts := cache.WriteOptions{}
	for _, o := range opts {
		o(&writeOpts)
	}

	key = m.prefix(key)

	m.cache.Set(key, value, writeOpts.Expiry)
	return nil
}

func (m *MemoryCache) Delete(key string, opts ...cache.DeleteOption) error {
	deleteOptions := cache.DeleteOptions{}
	for _, o := range opts {
		o(&deleteOptions)
	}

	m.cache.Delete(m.prefix(key))
	return nil
}
