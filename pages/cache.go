package pages

import (
	"log"
	"time"

	"github.com/karlseguin/ccache"
)

type Cache struct {
	lru *ccache.Cache
}

func NewCache() (*Cache, error) {
	lru := ccache.New(ccache.Configure().MaxSize(100).ItemsToPrune(25))

	return &Cache{
		lru: lru,
	}, nil
}

func (c *Cache) Get(key string) (*Page, error) {
	item := c.lru.Get(key)
	if item != nil { // Found an item in the cache
		log.Print("cache hit")
		return item.Value().(*Page), nil
	}

	log.Print("cache miss")

	page, err := BuildPage(key)
	if err != nil { // 500 Internal server error case
		return nil, err
	}

	if page == nil { // 404 Not Found case
		return nil, nil
	}

	// Add item to cache
	c.lru.Set(key, page, time.Hour*24*7)

	return page, nil
}
