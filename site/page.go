package site

import (
	"bytes"
	"text/template"
	"time"

	"github.com/antchfx/htmlquery"
	//log "github.com/sirupsen/logrus"
	"gopkg.in/russross/blackfriday.v2"
)

const numRecent = 6
const indexKey = "index"
const rssLimit = 20
const rssKey = "rss.xml"

func LoadPages(dir string, templates *template.Template, cache *ContentCache) error {
	err := buildMarkdownFiles(dir, templates, cache)
	if err != nil {
		return err
	}

	// Build index/home
	err = buildIndex(templates, cache)
	if err != nil {
		return err
	}

	// Build RSS
	err = buildRss(templates, cache)
	if err != nil {
		return err
	}

	return nil
}

func buildMarkdownFiles(dir string, templates *template.Template, cache *ContentCache) error {
	keys, err := getKeys(dir, ".md")
	if err != nil {
		return err
	}

	for _, key := range keys {
		page, err := buildPage(dir+key, templates)
		if err != nil {
			return err
		}

		cache.Set(dir+key, content)
	}

	return nil
}

func buildPage(key string, templates *template.Template) (*Content, error) {
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

	// Parse in to something we can query with xpath
	doc, err := htmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Get details from parsed html
	posts := p.posts.GetRecent(numRecent)
	title := getTitle(doc)

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
		return nil, err
	}

	content := buf.Bytes()

	p.cache.Set(key, &Page{
		Content:      &content,
		Mime:         "text/html; charset=utf-8",
		CacheControl: "public, must-revalidate",
		Etag:         getEtag(&content),
	})

	return nil, nil
}

func buildIndex(templates *template.Template) error {
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

func buildRss(templates *template.Template) error {
	// Build index/home
	posts := p.posts.GetRecent(rssLimit)

	// Run markdown through page template
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
