package site

import (
	"fmt"
	"time"

	"text/template"
)

type TemplateData struct {
	Key        string
	Title      string
	JavaScript string
	CSS        string
	Content    string
	Site       *Site
	Posts      *[]*Post
	Generated  time.Time
	Social     *Social
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
			return date.In(utc).Format(time.RFC822)
		},
		"GetAssetURL": func(key string, hashes Hashes) string {
			return fmt.Sprintf("/static/%s?m=%s", key, hashes[key])
		},
	})

	tmpl, err = tmpl.ParseGlob(templateDir + "*.tmpl")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
