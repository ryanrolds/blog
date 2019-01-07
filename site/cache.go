package site

import (
	log "github.com/sirupsen/logrus"
)

type Content struct {
	Content      *[]byte
	Etag         string
	Mime         string
	CacheControl string
}

type ContentCache map[string]*Content

type Cache struct {
	cache *ContentCache
}

func NewCache() *Cache {
	return &Cache{
		cache: &ContentCache{},
	}
}

func (c *Cache) GetKeys() []string {
	var keys []string
	for key := range *c.cache {
		keys = append(keys, key)
	}

	return keys
}

func (c *Cache) GetValues() []*Content {
	var values []*Content
	for _, value := range *c.cache {
		values = append(values, value)
	}

	return values
}

func (c *Cache) Get(key string) *Content {
	item, exists := (*c.cache)[key]
	if exists { // Found an item in the cache
		log.Debug("cache hit")
		return item
	}

	log.Debug("cache miss/stale")

	return item
}

func (c *Cache) Set(key string, item *Content) {
	(*c.cache)[key] = item
}

func (c *Cache) GetHashes() *Hashes {
	hashes := Hashes{}

	keys := c.GetKeys()
	for _, key := range keys {
		value := c.Get(key)
		hashes[key] = value.Etag
	}

	return &hashes
}
