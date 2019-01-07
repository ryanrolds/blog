package site

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func getKeys(dir string, suffix string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			key := strings.TrimSuffix(file.Name(), suffix)
			keys = append(keys, key)
		}
	}

	return keys, nil
}

func getAsset(filename string) (*[]byte, string, error) {
	// Get file contents
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, "", err
	}

	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)

	return &contents, mimeType, nil
}

func getCSS(key string) (*[]byte, error) {
	// Get file contents
	css, err := ioutil.ReadFile(key + ".css")
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
	javaScript, err := ioutil.ReadFile(key + ".js")
	if err != nil {
		if os.IsNotExist(err) {
			return &[]byte{}, nil
		}

		return nil, err
	}

	return &javaScript, nil
}

func getMarkdown(key string) (*[]byte, error) {
	// Get file contents
	log.Info("Loading file ", key+".md")
	content, err := ioutil.ReadFile(key + ".md")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	return &content, nil
}

func getEtag(buffer *[]byte) string {
	hash := md5.Sum(*buffer)
	return fmt.Sprintf("%x", hash)
}

func isPublished(doc *html.Node) bool {
	publishedAtElm := htmlquery.FindOne(doc, "//div[@id='published-at']")
	if publishedAtElm != nil {
		return true
	}

	return false
}

func getPublishedAt(doc *html.Node) time.Time {
	publishedAt := time.Now()
	publishedAtElm := htmlquery.FindOne(doc, "//div[@id='published-at']")
	if publishedAtElm != nil {
		publishedAtValue := htmlquery.InnerText(publishedAtElm)
		publishedAtParsed, err := time.Parse(time.RFC3339, publishedAtValue)
		if err != nil {
			log.Error(err)
		} else {
			publishedAt = publishedAtParsed
		}
	} else {
		log.Warnf("Published At not found for post")
	}

	return publishedAt
}

func getTitle(doc *html.Node) string {
	title := "Title"
	titleElm := htmlquery.FindOne(doc, "//h1")
	if titleElm != nil {
		title = htmlquery.InnerText(titleElm)
	} else {
		log.Warn("Title not found for post")
	}

	return html.EscapeString(title)
}

func getIntro(doc *html.Node) string {
	intro := "Intro"
	introElm := htmlquery.FindOne(doc, "//p")
	if introElm != nil {
		intro = htmlquery.InnerText(introElm)
	} else {
		log.Warn("Intro not found for post")
	}

	return intro
}

func getImage(doc *html.Node) string {
	image := ""
	imageElm := htmlquery.FindOne(doc, "//img")
	if imageElm != nil {
		image = htmlquery.SelectAttr(imageElm, "src")
	} else {
		log.Warn("Image not found for post")
	}

	return image
}
