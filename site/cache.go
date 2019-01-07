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

type ContentCache map[string]Content

func NewContentCache() *ContentCache {
	return &ContentCache{}
}

func (c *ContentCache) GetKeys() []string {
	var keys []string
	for key := range c {
		keys = append(keys, key)
	}

	return keys
}

func (c *ContentCache) GetValues() []Content {
	var values []interface{}
	for _, value := range c {
		values = append(values, value)
	}

	return values
}

func (c *ContentCache) Get(key string) Content {
	item, exists := c[key]
	if exists { // Found an item in the cache
		log.Debug("cache hit")
		return item
	}

	log.Debug("cache miss/stale")

	return item
}

func (c *ContentCache) Set(key string, item Content) {
	c[key] = item
}

func (c *ContentCache) GetHashes() *Hashes {
	hashes := Hashes{}

	keys := c.GetKeys()
	for _, key := range keys {
		value := c.Get(key)
		hashes[key] = value.(*Asset).Etag
	}

	return &hashes
}
