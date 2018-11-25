package pages

import (
	"bytes"
	"io/ioutil"
	"os"
	"text/template"

	"gopkg.in/russross/blackfriday.v2"
)

const (
	ContentDir   = "./content/"
	PagesDir     = ContentDir + "pages/"
	PostsDir     = ContentDir + "posts/"
	StaticDir    = ContentDir + "static/"
	TemplateFile = ContentDir + "template.html"
	IndexFile    = PagesDir + "index.md"
)

type TemplateDetails struct {
	JavaScript string
	CSS        string
	Body       string
}

type Page struct {
	Content *[]byte
}

func BuildPage(key string) (*Page, error) {
	markdown, err := getContent(key)
	if err != nil {
		return nil, err
	}

	// Page does not exist
	if markdown == nil {
		return nil, nil
	}

	template, err := getTemplate(key)
	if err != nil {
		return nil, err
	}

	css, err := getCSS(key)
	if err != nil {
		return nil, err
	}

	javaScript, err := getJavaScript(key)
	if err != nil {
		return nil, err
	}

	// Process MD
	body := blackfriday.Run(*markdown)

	buf := &bytes.Buffer{}
	err = template.Execute(buf, &TemplateDetails{
		CSS:        string((*css)[:]),
		JavaScript: string((*javaScript)[:]),
		Body:       string(body[:]),
	})
	if err != nil {
		return nil, err
	}

	content := buf.Bytes()

	return &Page{
		Content: &content,
	}, nil
}

func getTemplate(key string) (*template.Template, error) {
	tmpl, err := template.ParseFiles(ContentDir + key + ".tmpl")
	if err != nil {
		if key == "index" {
			return nil, err
		}

		tmpl, err = template.ParseFiles(ContentDir + "index.tmpl")
		if err != nil {
			return nil, err
		}
	}

	return tmpl, nil
}

func getCSS(key string) (*[]byte, error) {
	// Get file contents
	css, err := ioutil.ReadFile(ContentDir + key + ".css")
	if err != nil {
		if os.IsNotExist(err) {
			return &[]byte{}, nil
		}

		return nil, err
	}

	return &css, nil
}

func getJavaScript(key string) (*[]byte, error) {
	// Get file contents
	javaScript, err := ioutil.ReadFile(ContentDir + key + ".js")
	if err != nil {
		if os.IsNotExist(err) {
			return &[]byte{}, nil
		}

		return nil, err
	}

	return &javaScript, nil
}

func getContent(key string) (*[]byte, error) {
	// Get file contents
	content, err := ioutil.ReadFile(ContentDir + key + ".md")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	return &content, nil
}
