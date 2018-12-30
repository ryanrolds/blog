package site

import (
	"bytes"
	"text/template"
	"time"

	//log "github.com/sirupsen/logrus"
	"gopkg.in/russross/blackfriday.v2"
)

const numRecent = 6
const indexKey = "index"

type Page struct {
	Content *[]byte
	Etag    string
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

	// Build index/home
	posts := p.posts.GetRecent(numRecent)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(buf, "index.tmpl", &PageTemplate{
		CSS:        "",
		JavaScript: "",
		Content:    "",
		Posts:      &posts,
		Site:       p.site,
		Generated:  time.Now(),
	})
	if err != nil {
		return err
	}

	body := buf.Bytes()

	p.cache.Set(indexKey, &Page{
		Content: &body,
		Etag:    getEtag(&body),
	})

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
	Content    string
	Posts      *[]*Post
	Site       *Site
	Generated  time.Time
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
	posts := p.posts.GetRecent(numRecent)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(buf, "page.tmpl", &PageTemplate{
		CSS:        string((*css)[:]),
		JavaScript: string((*javaScript)[:]),
		Content:    string(body[:]),
		Posts:      &posts,
		Site:       p.site,
		Generated:  time.Now(),
	})
	if err != nil {
		return nil, err
	}

	content := buf.Bytes()

	return &Page{
		Content: &content,
	}, nil
}
