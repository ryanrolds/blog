package site

import (
	"bytes"
	"errors"
	"sort"
	"time"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	bf "gopkg.in/russross/blackfriday.v2"
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

func LoadPost() error {
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

func buildPost(key string) (*Post, error) {
	markdown, err := getMarkdown(p.dir + key)
	if err != nil {
		return nil, err
	}

	// Page does not exist
	if markdown == nil {
		return nil, nil
	}

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

	css := new(bytes.Buffer)
	if err = renderer.Formatter.WriteCSS(css, renderer.Style); err != nil {
		log.WithError(err).Warning("Couldn't write CSS")
	}

	// Process MD
	body := bf.Run(*markdown, bf.WithRenderer(renderer), bf.WithExtensions(exts))

	// Parse into something we can query with xpath
	doc, err := htmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// If no published date, skip
	if isPublished(doc) == false {
		log.Infof("Skipping %s, not published", key)
		return nil, ErrNotPublished
	}

	// Get details from parsed html
	title := getTitle(doc)
	publishedAt := getPublishedAt(doc)
	intro := getIntro(doc)
	image := getImage(doc)
	url := getPostUrl(p.site.Env, key)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(buf, "post.tmpl", &TemplateData{
		Key:        key,
		Title:      title,
		CSS:        css.String(),
		JavaScript: "",
		Content:    string(body[:]),
		Site:       p.site,
		Generated:  time.Now(),

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
