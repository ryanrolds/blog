package site

import (
	"time"

	"text/template"
)

func LoadTemplates(templateDir string) (*template.Template, error) {
	utc, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}

	tmpl := template.New("").Funcs(template.FuncMap{
		"FormatDate": func(date time.Time) string {
			return date.In(utc).Format(time.RFC3339)
		},
	})

	tmpl, err = tmpl.ParseGlob(templateDir + "*.tmpl")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
