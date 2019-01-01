package site

import (
	"bytes"
	"sort"
	"text/template"
	"time"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	"gopkg.in/russross/blackfriday.v2"
)

type Post struct {
	Slug      string
	Title     string
	Intro     string
	Image     string
	Content   *[]byte
	CreatedAt time.Time
	UpdatedAt time.Time
	Etag      string
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
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
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
	Content    string
	Site       *Site
	Generated  time.Time
}

func (p *PostManager) buildPost(key string) (*Post, error) {
	markdown, err := getMarkdown(p.dir + key)
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
	title := getTitle(doc)
	createdAt := getCreatedAt(doc)
	intro := getIntro(doc)
	image := getImage(doc)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(buf, "post.tmpl", &PostTemplate{
		CSS:        string((*css)[:]),
		JavaScript: string((*javaScript)[:]),
		Content:    string(body[:]),
		Site:       p.site,
		Generated:  time.Now(),
	})
	if err != nil {
		return nil, err
	}

	content := buf.Bytes()

	return &Post{
		Slug:      key,
		Title:     title,
		Image:     image,
		Intro:     intro,
		CreatedAt: createdAt,
		Content:   &content,
		Etag:      getEtag(&content),
	}, nil
}

func getCreatedAt(doc *html.Node) time.Time {
	createdAt := time.Now()
	createdAtElm := htmlquery.FindOne(doc, "//div[@id='created-at']")
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
	titleElm := htmlquery.FindOne(doc, "//h1")
	if titleElm != nil {
		title = htmlquery.InnerText(titleElm)
	} else {
		log.Warn("Title not found for post")
	}

	return html.EscapeString(title)
}

func getIntro(doc *html.Node) string {
	intro := "Intro"
	introElm := htmlquery.FindOne(doc, "//p")
	if introElm != nil {
		log.Info(introElm)
		intro = htmlquery.InnerText(introElm)
	} else {
		log.Warn("Intro not found for post")
	}

	return intro
}

func getImage(doc *html.Node) string {
	image := ""
	imageElm := htmlquery.FindOne(doc, "//img")
	if imageElm != nil {
		image = htmlquery.SelectAttr(imageElm, "src")
	} else {
		log.Warn("Image not found for post")
	}

	return image
}
