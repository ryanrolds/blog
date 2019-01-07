package site

import (
	"fmt"
	"text/template"
	"time"
)

type TemplateData struct {
	Key string

	Title       string
	Description string
	ImageUrl    string
	Url         string

	JavaScript string
	CSS        string
	Content    string
	Generated  time.Time

	Site  *Site
	Posts *PostList
}

func LoadTemplates(site *Site) (*template.Template, error) {
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}

	tmpl := template.New("").Funcs(template.FuncMap{
		"FormatDate": func(date time.Time) string {
			return date.In(utc).Format(time.RFC3339)
		},
		"GetAssetURL": func(key string, hashes Hashes) string {
			return fmt.Sprintf("/static/%s?m=%s", key, hashes["static/"+key])
		},
	})

	tmpl, err = tmpl.ParseGlob(site.rootDir + "*.tmpl")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
