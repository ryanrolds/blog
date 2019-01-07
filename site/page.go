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

func LoadPages(site *Site) error {
	err := buildMarkdownFiles(site, site.rootDir, site.templates, site.posts, site.cache)
	if err != nil {
		return err
	}

	// Build index/home
	err = buildIndex(site, site.templates, site.posts, site.cache)
	if err != nil {
		return err
	}

	// Build RSS
	err = buildRss(site, site.templates, site.posts, site.cache)
	if err != nil {
		return err
	}

	return nil
}

func buildMarkdownFiles(site *Site, dir string, templates *template.Template,
	posts *PostList, cache *Cache) error {
	keys, err := getKeys(dir, ".md")
	if err != nil {
		return err
	}

	for _, key := range keys {
		err := buildPage(site, key, templates, posts, cache)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildPage(site *Site, key string, templates *template.Template, posts *PostList,
	cache *Cache) error {
	markdown, err := getMarkdown(key)
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
	body := blackfriday.Run(*markdown)

	// Parse in to something we can query with xpath
	doc, err := htmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return err
	}

	// Get details from parsed html
	recent := (*posts)[:]
	if len(*posts) < numRecent {
		recent = recent[:len(recent)]
	} else {
		recent = recent[:numRecent]
	}

	title := getTitle(doc)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = templates.ExecuteTemplate(buf, "page.tmpl", &TemplateData{
		Title:      title,
		CSS:        string((*css)[:]),
		JavaScript: string((*javaScript)[:]),
		Content:    string(body[:]),
		Posts:      &recent,
		Site:       site,
		Generated:  time.Now(),
	})
	if err != nil {
		return err
	}

	content := buf.Bytes()

	cache.Set(key, &Content{
		Content:      &content,
		Mime:         "text/html; charset=utf-8",
		CacheControl: "public, must-revalidate",
		Etag:         getEtag(&content),
	})

	return nil
}

func buildIndex(site *Site, templates *template.Template, posts *PostList, cache *Cache) error {
	// Build index/home
	recent := (*posts)[:]
	if len(*posts) < numRecent {
		recent = recent[:len(recent)]
	} else {
		recent = recent[:numRecent]
	}

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err := templates.ExecuteTemplate(buf, "index.tmpl", &TemplateData{
		Title:      "Home",
		CSS:        "",
		JavaScript: "",
		Content:    "",
		Posts:      &recent,
		Site:       site,
		Generated:  time.Now(),
	})
	if err != nil {
		return err
	}

	body := buf.Bytes()

	cache.Set(indexKey, &Content{
		Content:      &body,
		Etag:         getEtag(&body),
		Mime:         "text/html; charset=utf-8",
		CacheControl: "public, must-revalidate",
	})

	return nil
}

func buildRss(site *Site, templates *template.Template, posts *PostList, cache *Cache) error {
	// Build index/home
	recent := (*posts)[:]
	if len(*posts) < numRecent {
		recent = recent[:len(recent)]
	} else {
		recent = recent[:rssLimit]
	}

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err := templates.ExecuteTemplate(buf, "rss.tmpl", &TemplateData{
		Title:      "",
		CSS:        "",
		JavaScript: "",
		Content:    "",
		Posts:      &recent,
		Site:       site,
		Generated:  time.Now(),
	})
	if err != nil {
		return err
	}

	body := buf.Bytes()

	cache.Set(rssKey, &Content{
		Content:      &body,
		Etag:         getEtag(&body),
		Mime:         "application/rss+xml; charset=utf-8",
		CacheControl: "public, must-revalidate",
	})

	return nil
}
