package site

import (
	"bytes"
	"sort"
	"text/template"
	"time"

	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"gopkg.in/russross/blackfriday.v2"
)

type Post struct {
	Title     string
	CreatedAt time.Time
	Content   *[]byte
}

type PostManager struct {
	dir         string
	templates   *template.Template
	cache       *Cache
	orderedList []*Post
	site        *Site
}

func NewPostManager(site *Site, dir string, templates *template.Template) *PostManager {
	return &PostManager{
		dir:       dir,
		templates: templates,
		cache:     NewCache(),
		site:      site,
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

	values := p.cache.GetValues()

	posts := []*Post{}
	for _, post := range values {
		posts = append(posts, post.(*Post))
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.Before(posts[j].CreatedAt)
	})

	p.orderedList = posts

	return nil
}

func (p *PostManager) Get(key string) *Post {
	item := p.cache.Get(key)
	if item == nil {
		return nil
	}

	return item.(*Post)
}

func (p *PostManager) GetRecent(num int) []*Post {
	if num > len(p.orderedList) {
		num = len(p.orderedList)
	}

	return p.orderedList[:num]
}

type PostTemplate struct {
	JavaScript string
	CSS        string
	Body       string
	Site       *Site
}

func (p *PostManager) buildPost(key string) (*Post, error) {
	markdown, err := getMarkdown(p.dir + key)
	if err != nil {
		return nil, err
	}

	log.Debug(markdown)

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

	doc, err := htmlquery.Parse(bytes.NewReader(body))

	title := getTitle(doc)
	createdAt := getCreatedAt(doc)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(buf, "post.tmpl", &PostTemplate{
		CSS:        string((*css)[:]),
		JavaScript: string((*javaScript)[:]),
		Body:       string(body[:]),
		Site:       p.site,
	})
	if err != nil {
		return nil, err
	}

	content := buf.Bytes()

	return &Post{
		Title:     title,
		CreatedAt: createdAt,
		Content:   &content,
	}, nil
}

func getCreatedAt(doc *html.Node) time.Time {
	createdAt := time.Now()
	createdAtElm := htmlquery.FindOne(doc, "//p/span[@id='created_at']")
	if createdAtElm != nil {
		createdAtValue := htmlquery.InnerText(createdAtElm)
		createdAtParsed, err := time.Parse(time.RFC3339, createdAtValue)
		if err != nil {
			log.Error(err)
		} else {
			createdAt = createdAtParsed
		}
	} else {
		log.Warnf("Created At not found for post")
	}

	return createdAt
}

func getTitle(doc *html.Node) string {
	title := "Title"
	titleElm := htmlquery.FindOne(doc, "//h1[0]")
	if titleElm != nil {
		title = htmlquery.InnerText(titleElm)
	} else {
		log.Warn("Title not found for post")
	}

	return title
}
