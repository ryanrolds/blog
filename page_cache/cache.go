package page_cache

import (
  "bytes"
  "fmt"
  "io/ioutil"
  "log"
  "text/template"
  "time"


  "github.com/karlseguin/ccache"
  "github.com/ryanrolds/pedantic_orderliness/types"
  "gopkg.in/russross/blackfriday.v2"
)

type TemplateDetails struct {
  Body string
}

type Cache struct {
  lru *ccache.Cache
  templates *template.Template
}

func NewCache() (*Cache, error) {
  templates, err := template.ParseFiles("./site/template.html")
  if err != nil {
    log.Print("Unable to parse template files")
    return nil, err
  }

  log.Print(templates.DefinedTemplates())

  lru := ccache.New(ccache.Configure().MaxSize(100).ItemsToPrune(25))

  return &Cache{
    templates: templates,
    lru: lru,
  }, nil
}


func (c *Cache) Get(key string) *types.Page {
  item := c.lru.Get(key)
  if item == nil {
    // Get file contents
    markdown, err := ioutil.ReadFile(fmt.Sprintf("./site/content%s.md", key))
    if err != nil {
      log.Print(err)
      return nil
    }

    // Process MD
    body := blackfriday.Run(markdown)

    buf := &bytes.Buffer{}
    err = c.templates.ExecuteTemplate(buf, "template.html", &TemplateDetails{
      Body: string(body[:]),
    })
    if err != nil {
      log.Print("Problem executing template")
      log.Print(err)
      return nil
    }

    content := buf.Bytes()

    // Cache the content
    page := &types.Page{
      Content: &content,
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
