package site

import (
	"bytes"
	"text/template"

	//log "github.com/sirupsen/logrus"
	"gopkg.in/russross/blackfriday.v2"
)

const numRecent = 6

type Page struct {
	Content *[]byte
}

type PageManager struct {
	dir       string
	templates *template.Template
	cache     *Cache
	posts     *PostManager
	site      *Site
}

func NewPageManager(site *Site, dir string, templates *template.Template, posts *PostManager) *PageManager {
	return &PageManager{
		dir:       dir,
		templates: templates,
		cache:     NewCache(),
		posts:     posts,
		site:      site,
	}
}

func (p *PageManager) Load() error {
	keys, err := getKeys(p.dir, ".md")
	if err != nil {
		return err
	}

	for _, key := range keys {
		page, err := p.buildPage(p.dir + key)
		if err != nil {
			return err
		}

		p.cache.Set(key, page)
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
	Posts      []*Post
	Site       *Site
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
		Posts:      p.posts.GetRecent(numRecent),
		Site:       p.site,
	})
	if err != nil {
		return nil, err
	}

	content := buf.Bytes()

	return &Page{
		Content: &content,
	}, nil
}
