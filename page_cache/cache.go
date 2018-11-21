package page_cache

import (
  "bytes"
  "fmt"
  "io/ioutil"
  "log"
  "time"


  "github.com/karlseguin/ccache"
  "github.com/ryanrolds/pedantic_orderliness/types"
  "gopkg.in/russross/blackfriday.v2"
)

func NewCache() *Cache {
  lru := ccache.New(ccache.Configure().MaxSize(100).ItemsToPrune(25))

  return &Cache{
    lru: lru,
  }
}

type Cache struct {
  lru *ccache.Cache
}

func (c *Cache) Get(key string) *types.Page {
  item := c.lru.Get(key)
  if item == nil {
    // Get file contents
    content, err := ioutil.ReadFile(fmt.Sprintf("./pages%s.md", key))
    if err != nil {
      log.Print(err)
      return nil
    }

    // Process MD
    output := blackfriday.Run(content)

    // Cache the content
    page := &types.Page{
      Content: bytes.NewBuffer(output),
    }
    c.lru.Set(key, page, time.Hour * 24 * 7)

    return page
  }

  page := item.Value().(*types.Page)
  return page
}

func (c *Cache) Get404() *types.Page {
  page := c.Get("/404")
  if page == nil {
    log.Panic("Unable to get 404 page")
  }

  return page
}
