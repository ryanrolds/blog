package site

import (
	"io/ioutil"
	"os"
	"strings"
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

func getAsset(filename string) (*[]byte, error) {
	// Get file contents
	contents, err := ioutil.ReadFile(AssetsDir + filename)
	if err != nil {
		return nil, err
	}

	return &contents, nil
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

func getMarkdown(key string) (*[]byte, error) {
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
