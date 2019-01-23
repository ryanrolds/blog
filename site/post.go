package site

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"text/template"
	"time"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	bf "gopkg.in/russross/blackfriday.v2"
)

var ErrNotPublished = errors.New("Not published")

type Post struct {
	Key         string
	Title       string
	Description string
	Image       string
	Content     *[]byte
	Amp         *[]byte
	PublishedAt time.Time
	UpdatedAt   time.Time
	Etag        string
	Url         string
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
			if err == ErrNotPublished {
				continue
			}

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
		return posts[i].PublishedAt.After(posts[j].PublishedAt)
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

func (p *PostManager) buildPost(key string) (*Post, error) {
	markdown, err := getMarkdown(p.dir+key, p.site.Log)
	if err != nil {
		return nil, err
	}

	// Page does not exist
	if markdown == nil {
		return nil, nil
	}

	body, css, err := renderMarkdown(markdown)
	if err != nil {
		return nil, err
	}

	// Parse into something we can query with xpath
	doc, err := htmlquery.Parse(bytes.NewReader(*body))
	if err != nil {
		return nil, err
	}

	// If no published date, skip
	if isPublished(doc) == false && p.site.Env == "production" {
		log.Infof("Skipping %s, not published", key)
		return nil, ErrNotPublished
	}

	// Get details from parsed html
	title := getTitle(doc, p.site.Log)
	publishedAt := getPublishedAt(doc, p.site.Log)
	description := getDescription(doc, p.site.Log)
	image := getImage(doc, p.site.Log)
	url := getPostUrl(p.site.Env, key)

	tmplData := &TemplateData{
		Key:         key,
		Title:       title,
		CSS:         css.String(),
		JavaScript:  "",
		Content:     string((*body)[:]),
		Site:        p.site,
		Generated:   time.Now(),
		PublishedAt: publishedAt,
		Social: &Social{
			Title:       title,
			Description: description,
			ImageUrl:    image,
			Url:         url,
		},
	}

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(buf, "post.tmpl", tmplData)
	if err != nil {
		return nil, err
	}

	content := buf.Bytes()

	// Run markdown through amp template
	ampBuf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(ampBuf, "amp.tmpl", tmplData)
	if err != nil {
		return nil, err
	}

	amp := ampBuf.Bytes()

	return &Post{
		Key:         key,
		Title:       title,
		Image:       image,
		Description: description,
		PublishedAt: publishedAt,
		Content:     &content,
		Amp:         &amp,
		Etag:        getEtag(&content),
		Url:         url,
	}, nil
}

func renderMarkdown(markdown *[]byte) (*[]byte, *bytes.Buffer, error) {
	// Defines the extensions that are used
	var exts = bf.NoIntraEmphasis | bf.Tables | bf.FencedCode | bf.Autolink |
		bf.Strikethrough | bf.SpaceHeadings | bf.BackslashLineBreak |
		bf.DefinitionLists | bf.Footnotes

	// Defines the HTML rendering flags that are used
	var flags = bf.UseXHTML | bf.Smartypants | bf.SmartypantsFractions |
		bf.SmartypantsDashes | bf.SmartypantsLatexDashes

	// Setting chroma renderer
	renderer := bfchroma.NewRenderer(
		bfchroma.Style("emacs"),
		bfchroma.WithoutAutodetect(),
		bfchroma.ChromaOptions(
			html.WithLineNumbers(),
			html.WithClasses(),
		),
		bfchroma.Extend(
			bf.NewHTMLRenderer(bf.HTMLRendererParameters{
				Flags: flags,
			}),
		),
	)

	css := bytes.Buffer{}
	if err := renderer.Formatter.WriteCSS(&css, renderer.Style); err != nil {
		log.WithError(err).Warning("Couldn't write CSS")
		return nil, nil, err
	}

	// Process MD
	body := bf.Run(*markdown, bf.WithRenderer(renderer), bf.WithExtensions(exts))

	return &body, &css, nil
}

func getPostUrl(env string, key string) string {
	domain := "test.pedanticorderliness.com"
	if env == "production" {
		domain = "www.pedanticorderliness.com"
	} else if env == "test" {
		domain = "test.pedanticorderliness.com"
	}

	return fmt.Sprintf("https://%s/posts/%s", domain, key)
}
