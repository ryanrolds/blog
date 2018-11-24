package pages

import (
	"log"
	"time"

	"github.com/karlseguin/ccache"
)

type Cache struct {
	lru *ccache.Cache
	ttl time.Duration
}

func NewCache(env string) (*Cache, error) {
	lru := ccache.New(ccache.Configure().MaxSize(100).ItemsToPrune(25))

	ttl := time.Hour * 24 * 7
	if env != "production" {
		ttl = time.Second * 5
	}

	log.Print("TTL: " + ttl.String())

	return &Cache{
		lru: lru,
		ttl: ttl,
	}, nil
}

func (c *Cache) Get(key string) (*Page, error) {
	item := c.lru.Get(key)
	if item != nil { // Found an item in the cache
		if item.TTL() > 0 {
			log.Print("cache hit")
			return item.Value().(*Page), nil
		}
	}

	log.Print("cache miss/stale")

	page, err := BuildPage(key)
	if err != nil { // 500 Internal server error case
		return nil, err
	}

	if page == nil { // 404 Not Found case
		return nil, nil
	}

	// Add item to cache
	c.lru.Set(key, page, c.ttl)

	return page, nil
}
