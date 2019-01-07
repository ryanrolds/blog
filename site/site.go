package site

import (
	"fmt"
	"net/http"
	//"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	postsDir  = "posts/"
	assetsDir = "static/"
)

type Hashes map[string]string

type Site struct {
	port    string
	Env     string
	rootDir string

	cache     *Cache
	templates *template.Template
	posts     *PostList
	Hashes    *Hashes
}

func NewSite(port, env, contentDir string) *Site {
	return &Site{
		port:    port,
		Env:     env,
		rootDir: contentDir,
	}
}

func (s *Site) Run() error {
	s.cache = NewCache()

	err := LoadAssets(s, assetsDir)
	if err != nil {
		return err
	}

	s.Hashes = s.cache.GetHashes()

	s.templates, err = LoadTemplates(s)
	if err != nil {
		return err
	}

	s.posts, err = LoadPosts(s, postsDir)
	if err != nil {
		return err
	}

	err = LoadPages(s)
	if err != nil {
		return err
	}

	// Prepare server
	server := http.Server{
		Addr:         fmt.Sprintf(":%s", s.port),
		Handler:      s,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info("Starting server")

	// Run server and block
	err = server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (s *Site) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// The root page uses the "index" key
	if path == "" || path == "/" {
		path = "/index"
	}

	// Favicon
	if path == "/favicon.ico" {
		path = "/static/favicon.ico"
	}

	// Robots.txt
	if path == "/robots.txt" {
		if s.Env == "production" {
			path = "/static/allow.txt"
		} else {
			path = "/static/disallow.txt"
		}
	}

	// Try to get cached content
	item := s.cache.Get(path)
	if item == nil {
		s.Handle404(w, r)
		return
	}

	if r.Header.Get("If-None-Match") == item.Etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", item.Mime)
	w.Header().Set("Cache-Control", item.CacheControl)
	//w.Header().Set("Cache-Control", "public, must-revalidate")
	//w.Header().Set("Cache-Control", "public, max-age=2419200")
	//w.Header().Set("Cache-Control", "public, max-age=604800")
	w.Header().Set("Etag", item.Etag)
	w.WriteHeader(http.StatusOK)
	w.Write(*item.Content)
}

func (s *Site) Handle404(w http.ResponseWriter, r *http.Request) {
	item := s.cache.Get("/404")
	if item == nil {
		s.Handle500(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusNotFound)
	w.Write(*item.Content)
}

func (s *Site) Handle500(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)

	item := s.cache.Get("/500")
	if item == nil {
		log.Warn("Unable to get 500 page")
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(*item.Content)
}

func getHost(env string) string {
	domain := "localhost:8080"
	if env == "production" {
		domain = "www.pedanticorderliness.com"
	} else if env == "test" {
		domain = "test.pedanticorderliness.com"
	}

	return domain
}
