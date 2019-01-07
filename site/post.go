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
	Intro       string
	Image       string
	Url         string
	PublishedAt time.Time
	UpdatedAt   time.Time
}

type PostList []*Post

func LoadPosts(site *Site, dir string) (*PostList, error) {
	keys, err := getKeys(site.rootDir+dir, ".md")
	if err != nil {
		return nil, err
	}

	posts := PostList{}

	for _, key := range keys {
		post, err := buildPost(site, key, site.templates, site.cache)
		if err != nil {
			if err == ErrNotPublished {
				continue
			}

			return nil, err
		}

		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].PublishedAt.After(posts[j].PublishedAt)
	})

	return &posts, nil
}

func buildPost(site *Site, key string, templates *template.Template, cache *Cache) (*Post, error) {
	markdown, err := getMarkdown(site.rootDir + postsDir + key)
	if err != nil {
		return nil, err
	}

	// Page does not exist
	if markdown == nil {
		return nil, errors.New("Missing Markdown file")
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
	url := getPostUrl(site, key)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = templates.ExecuteTemplate(buf, "post.tmpl", &TemplateData{
		Key:        key,
		Title:      title,
		CSS:        css.String(),
		JavaScript: "",
		Content:    string(body[:]),
		Site:       site,
		Generated:  time.Now(),
	})
	if err != nil {
		return nil, err
	}

	content := buf.Bytes()

	cache.Set("/posts/"+key, &Content{
		Content:      &content,
		Etag:         getEtag(&content),
		Mime:         "text/html; charset=utf-8",
		CacheControl: "public, must-revalidate",
	})

	return &Post{
		Key:         key,
		Title:       title,
		Image:       image,
		Intro:       intro,
		PublishedAt: publishedAt,
		Url:         url,
	}, nil
}

func getPostUrl(site *Site, key string) string {
	return fmt.Sprintf("https://%s/%s", getHost(site.Env), key)
}
