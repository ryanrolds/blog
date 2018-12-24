package site

import (
	"filepath"
	"io/ioutil"
	"mime"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
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
	contents, err := ioutil.ReadFile(AssetsDir + filename)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)

	log.Warn(filename, ext, mimeType)

	log.Error("asdfasdf", mimeType)

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
