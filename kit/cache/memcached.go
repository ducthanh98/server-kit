package cache

import (
	go_cache "github.com/patrickmn/go-cache"
	"time"
)

type Memcached struct {
	instance *go_cache.Cache
}

func NewMemcached() *Memcached {
	instance := go_cache.New(15*time.Minute, 5*time.Minute)

	return &Memcached{
		instance: instance,
	}
}

func (c *Memcached) Get(key string, callback Callback) (interface{}, error) {
	if result, found := c.instance.Get(key); found {
		return result, nil
	}

	data, err := callback()
	if err != nil {
		return nil, err
	}
	c.Set(key, data)
	return data, nil

}

func (c *Memcached) Set(key string, value interface{}) {
	c.instance.Set(key, value, go_cache.DefaultExpiration)
}

func (c *Memcached) Delete(key string) {
	c.instance.Delete(key)
}
