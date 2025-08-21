package site

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"text/template"
)

type TemplateData struct {
	Key         string
	Title       string
	JavaScript  string
	CSS         string
	Content     string
	Site        *Site
	Posts       *[]*Post
	Generated   time.Time
	PublishedAt time.Time
	Social      *Social
}

type Social struct {
	Title       string
	Description string
	ImageUrl    string
	Url         string
}

func LoadTemplates(templateDir string) (*template.Template, error) {
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}

	tmpl := template.New("").Funcs(template.FuncMap{
		"FormatDate": func(date time.Time) string {
			return date.In(utc).Format(time.RFC3339)
		},
		"FormatRssDate": func(date time.Time) string {
			return date.In(utc).Format(time.RFC1123Z)
		},
		"GetAssetURL": func(key string, hashes Hashes) string {
			return fmt.Sprintf("/static/%s?m=%s", key, hashes[key])
		},
	})

	err = fs.WalkDir(ContentFS, templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		
		content, err := fs.ReadFile(ContentFS, path)
		if err != nil {
			return err
		}
		
		name := filepath.Base(path)
		_, err = tmpl.New(name).Parse(string(content))
		return err
	})
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
