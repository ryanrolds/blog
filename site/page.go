package site

import (
	"bytes"
	"text/template"
	"time"

	"github.com/antchfx/htmlquery"
	bf "github.com/russross/blackfriday/v2"
)

const numRecent = 6
const indexKey = "index"
const rssLimit = 20
const rssKey = "rss.xml"

type Page struct {
	Content      *[]byte
	Mime         string
	Etag         string
	CacheControl string
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
	err := p.buildMarkdownFiles()
	if err != nil {
		return err
	}

	// Build index/home
	err = p.buildIndex()
	if err != nil {
		return err
	}

	// Build RSS
	err = p.buildRss()
	if err != nil {
		return err
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

func (p *PageManager) buildMarkdownFiles() error {
	keys, err := getKeys(p.dir, ".md")
	if err != nil {
		return err
	}

	for _, key := range keys {
		err := p.buildPage(p.dir + key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PageManager) buildPage(key string) error {
	markdown, err := getMarkdown(key, p.site.Log)
	if err != nil {
		return err
	}

	// Page does not exist
	if markdown == nil {
		return nil
	}

	css, err := getCSS(key)
	if err != nil {
		return err
	}

	javaScript, err := getJavaScript(key)
	if err != nil {
		return err
	}

	// Process MD
	body := bf.Run(*markdown)

	// Parse in to something we can query with xpath
	doc, err := htmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return err
	}

	// Get details from parsed html
	posts := p.posts.GetRecent(numRecent)
	title := getTitle(doc, p.site.Log)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(buf, "page.tmpl", &TemplateData{
		Title:      title,
		CSS:        string((*css)[:]),
		JavaScript: string((*javaScript)[:]),
		Content:    string(body[:]),
		Posts:      &posts,
		Site:       p.site,
		Social:     &Social{},
		Generated:  time.Now(),
	})
	if err != nil {
		return err
	}

	content := buf.Bytes()

	p.cache.Set(key, &Page{
		Content:      &content,
		Mime:         "text/html; charset=utf-8",
		CacheControl: "public, must-revalidate",
		Etag:         getEtag(&content),
	})

	return nil
}

func (p *PageManager) buildIndex() error {
	// Build index/home
	posts := p.posts.GetRecent(numRecent)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err := p.templates.ExecuteTemplate(buf, "index.tmpl", &TemplateData{
		Title:      "Home",
		CSS:        "",
		JavaScript: "",
		Content:    "",
		Posts:      &posts,
		Social:     &Social{},
		Site:       p.site,
		Generated:  time.Now(),
	})
	if err != nil {
		return err
	}

	body := buf.Bytes()

	p.cache.Set(indexKey, &Page{
		Content:      &body,
		Etag:         getEtag(&body),
		Mime:         "text/html; charset=utf-8",
		CacheControl: "public, must-revalidate",
	})

	return nil
}

func (p *PageManager) buildRss() error {
	// Get a list of most recent posts
	posts := p.posts.GetRecent(rssLimit)

	buf := &bytes.Buffer{}
	err := p.templates.ExecuteTemplate(buf, "rss.tmpl", &TemplateData{
		Title:      "",
		CSS:        "",
		JavaScript: "",
		Content:    "",
		Posts:      &posts,
		Social:     &Social{},
		Site:       p.site,
		Generated:  time.Now(),
	})
	if err != nil {
		return err
	}

	body := buf.Bytes()

	p.cache.Set(rssKey, &Page{
		Content:      &body,
		Etag:         getEtag(&body),
		Mime:         "application/rss+xml; charset=utf-8",
		CacheControl: "public, must-revalidate",
	})

	return nil
}
