package pages

import (
	"log"
)

type ContentCache map[string]*Page

type Cache struct {
	cache ContentCache
	env   string
}

func NewCache(env string) (*Cache, error) {
	return &Cache{
		cache: ContentCache{},
		env:   env,
	}, nil
}

func (c *Cache) Get(key string) (*Page, error) {
	item, exists := c.cache[key]
	if exists { // Found an item in the cache
		log.Print("cache hit")
		return item, nil
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
	if c.env == "production" {
		c.cache[key] = page
	}

	return page, nil
}
