package site

import (
	"bytes"
	"sort"
	"text/template"
	"time"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	bf "gopkg.in/russross/blackfriday.v2"
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
	Slug       string
	Title      string
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

	// Get details from parsed html
	title := getTitle(doc)
	createdAt := getCreatedAt(doc)
	intro := getIntro(doc)
	image := getImage(doc)

	// Run markdown through page template
	buf := &bytes.Buffer{}
	err = p.templates.ExecuteTemplate(buf, "post.tmpl", &PostTemplate{
		Slug:       key,
		Title:      title,
		CSS:        css.String(),
		JavaScript: "",
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
