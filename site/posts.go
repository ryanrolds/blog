package site

import (
	"bytes"
	"text/template"

	"gopkg.in/russross/blackfriday.v2"
)

type Post struct {
	Content *[]byte
}

type PostManager struct {
	dir       string
	templates *template.Template
	cache     *Cache
}

func NewPostManager(dir string, templates *template.Template) *PostManager {
	return &PostManager{
		dir:       dir,
		templates: templates,
		cache:     NewCache(),
	}
}

func (p *PostManager) Load() error {
	keys, err := getKeys(p.dir, ".md")
	if err != nil {
		return err
	}

	for _, key := range keys {
		post, err := p.buildPost(key)
		if err != nil {
			return err
		}

		p.cache.Set(key, post)
	}

	return nil
}

func (p *PostManager) Get(key string) *Post {
	item := p.cache.Get(key)
	return item.(*Post)
}

type PostTemplate struct {
	JavaScript string
	CSS        string
	Body       string
}

func (p *PostManager) buildPost(key string) (*Post, error) {
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
	err = p.templates.ExecuteTemplate(buf, "post.tmpl", &PostTemplate{
		CSS:        string((*css)[:]),
		JavaScript: string((*javaScript)[:]),
		Body:       string(body[:]),
	})
	if err != nil {
		return nil, err
	}

	content := buf.Bytes()

	return &Post{
		Content: &content,
	}, nil
}
