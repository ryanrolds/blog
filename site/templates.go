package site

import (
	"text/template"
)

func LoadTemplates(templateDir string) (*template.Template, error) {
	tmpl, err := template.ParseGlob(templateDir + "*.tmpl")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
