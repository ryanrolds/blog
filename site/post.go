package site

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"text/template"
	"time"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/gernest/front"

	"github.com/pkg/errors"
	bf "github.com/russross/blackfriday/v2"
	log "github.com/sirupsen/logrus"
)

var ErrNotPublished = errors.New("Not published")

type Post struct {
	Slug        string
	Title       string
	Intro       string
	Image       string
	Content     *[]byte
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
	matter      *front.Matter
}

func NewPostManager(site *Site, dir string, templates *template.Template) *PostManager {
	m := front.NewMatter()
	m.Handle("---", front.YAMLHandler)

	return &PostManager{
		dir:       dir,
		templates: templates,
		cache:     NewCache(),
		site:      site,
		matter:    m,
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
	filename := p.dir + key + ".md"
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "problem reading file %s", filename)
	}

	front, markdown, err := p.matter.Parse(file)
	if err != nil {
		return nil, errors.Wrapf(err, "problem parsing file %s", filename)
	}

	byteMarkdown := []byte(markdown)

	body, css, err := renderMarkdown(&byteMarkdown)
	if err != nil {
		return nil, err
	}

	publishedAt, err := getDateFromFrontMatter(front, "published")
	if err != nil {
		return nil, errors.Wrapf(err, "problem getting published date from %s", filename)
	}

	now := time.Now()

	// If no published date, skip
	if now.Before(publishedAt) && p.site.Env == "production" {
		log.Infof("Skipping %s, not published", key)
		return nil, ErrNotPublished
	}

	// Get details from parsed html
	title, err := getStringFromFrontMatter(front, "title")
	if err != nil {
		return nil, errors.Wrapf(err, "problem getting title from %s", filename)
	}

	intro, err := getStringFromFrontMatter(front, "intro")
	if err != nil {
		return nil, errors.Wrapf(err, "problem getting intro from %s", filename)
	}

	image, err := getStringFromFrontMatter(front, "image")
	if err != nil {
		log.Warnf("problem getting image from %s", filename)
	}

	url, err := getStringFromFrontMatter(front, "url")
	if err != nil {
		log.Warnf("problem getting url from %s", filename)
	}

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(buf, "post.tmpl", &TemplateData{
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
			Description: intro,
			ImageUrl:    image,
			Url:         url,
		},
	})
	if err != nil {
		return nil, err
	}

	content := buf.Bytes()

	return &Post{
		Slug:        key,
		Title:       title,
		Image:       image,
		Intro:       intro,
		PublishedAt: publishedAt,
		Content:     &content,
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
			html.WithLineNumbers(true),
			html.WithClasses(true),
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
