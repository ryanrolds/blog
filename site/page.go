package site

import (
	"bytes"
	"log"
	"text/template"

	"gopkg.in/russross/blackfriday.v2"
)

type Page struct {
	Content *[]byte
}

type PageManager struct {
	dir       string
	templates *template.Template
	cache     *Cache
}

func NewPageManager(dir string, templates *template.Template) *PageManager {
	return &PageManager{
		dir:       dir,
		templates: templates,
		cache:     NewCache(),
	}
}

func (p *PageManager) Load() error {
	keys, err := getKeys(p.dir, ".md")
	if err != nil {
		return err
	}

	for _, key := range keys {
		page, err := p.buildPage(key)
		if err != nil {
			return err
		}

		p.cache.Set(key, page)
	}

	keys = p.cache.GetKeys()
	for _, key := range keys {
		log.Print(key)
	}

	return nil
}

func (p *PageManager) Get(key string) *Page {
	item := p.cache.Get(key)
	if item == nil {
		return nil
	}

	return item.(*Page)
}

type PageTemplate struct {
	JavaScript string
	CSS        string
	Body       string
}

func (p *PageManager) buildPage(key string) (*Page, error) {
	markdown, err := getMarkdown(key)
	if err != nil {
		return nil, err
	}

	// Page does not exist
	if markdown == nil {
		return nil, nil
	}

	css, err := getCSS(key)
	if err != nil {
		return nil, err
	}

	javaScript, err := getJavaScript(key)
	if err != nil {
		return nil, err
	}

	// Process MD
	body := blackfriday.Run(*markdown)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(buf, "page.tmpl", &PageTemplate{
		CSS:        string((*css)[:]),
		JavaScript: string((*javaScript)[:]),
		Body:       string(body[:]),
	})
	if err != nil {
		return nil, err
	}

	content := buf.Bytes()

	return &Page{
		Content: &content,
	}, nil
}
