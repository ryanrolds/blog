package site

import (
	log "github.com/sirupsen/logrus"
)

type ContentCache map[string]interface{}

type Cache struct {
	cache ContentCache
}

func NewCache() *Cache {
	return &Cache{
		cache: ContentCache{},
	}
}

func (c *Cache) GetKeys() []string {
	var keys []string
	for key := range c.cache {
		keys = append(keys, key)
	}

	return keys
}

func (c *Cache) GetValues() []interface{} {
	var values []interface{}
	for _, value := range c.cache {
		values = append(values, value)
	}

	return values
}

func (c *Cache) Get(key string) interface{} {
	item, exists := c.cache[key]
	if exists { // Found an item in the cache
		log.Debug("cache hit")
		return item
	}

	log.Debug("cache miss/stale")

	return item
}

func (c *Cache) Set(key string, item interface{}) {
	c.cache[key] = item
}
