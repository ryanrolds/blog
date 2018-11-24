package site

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ryanrolds/pedantic_orderliness/site/pages"
)

type Site struct {
	Port  string
	Env   string
	cache *pages.Cache
}

func NewSite(port string, env string) (*Site, error) {
	cache, err := pages.NewCache(env)
	if err != nil {
		return nil, err
	}

	return &Site{
		Port:  port,
		Env:   env,
		cache: cache,
	}, nil
}

func (s *Site) Run() error {
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", s.Port),
		Handler: s,
	}

	// Blocks
	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (s *Site) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// We want the index file
	if path == "/" {
		path = "index"
	}

	// Help protect from reading files outside of the app
	path = strings.TrimLeft(path, "/")

	log.Print("Request: " + path)

	page, err := s.cache.Get(path)
	if err != nil { // Error getting the page
		page, err = s.cache.Get("500")
		if err != nil {
			log.Panic("Unable to get 500 page")
		}

		rw.WriteHeader(http.StatusInternalServerError)
	}

	// If we couldn't find a page, then redirect to 404
	if page == nil { // No page was found
		page, err = s.cache.Get("404")
		if err != nil {
			log.Panic("Unable to get 404 page")
		}

		rw.WriteHeader(http.StatusNotFound)
	} else {
		rw.WriteHeader(http.StatusOK)
	}

	rw.Header().Set("Content-Type", "text/html")
	rw.Write(*page.Content)
}
