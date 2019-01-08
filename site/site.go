package site

import (
	"net/http"
	//"strings"
	"text/template"

	"github.com/gorilla/mux"
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

func (s *Site) GetHandler() (*mux.Router, error) {
	err := s.loadContent()
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	r.HandleFunc("/static/{key}", s.assetHandler)
	r.HandleFunc("/posts/{key}", s.postHandler)
	r.HandleFunc("/robots.txt", s.robotsTxtHandler)
	r.HandleFunc("/rss.xml", s.rssHandler)
	r.HandleFunc("/favicon.ico", s.faviconHandler)
	r.HandleFunc("/{key}", s.pageHandler)
	r.HandleFunc("/", s.indexHandler)
	r.HandleFunc("", s.indexHandler)
	r.NotFoundHandler = http.HandlerFunc(s.handle404)

	return r, nil
}

func (s *Site) loadContent() error {
	s.cache = NewCache()

	err := loadAssets(s, assetsDir)
	if err != nil {
		return err
	}

	s.Hashes = s.cache.GetHashes()

	s.templates, err = loadTemplates(s)
	if err != nil {
		return err
	}

	s.posts, err = loadPosts(s, postsDir)
	if err != nil {
		return err
	}

	err = loadPages(s)
	if err != nil {
		return err
	}

	return nil
}

func (s *Site) assetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	s.cacheHandler(w, r, assetsDir+key)
}

func (s *Site) postHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	s.cacheHandler(w, r, postsDir+key)
}

func (s *Site) pageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	s.cacheHandler(w, r, postsDir+key)
}

func (s *Site) indexHandler(w http.ResponseWriter, r *http.Request) {
	s.cacheHandler(w, r, "index")
}

func (s *Site) rssHandler(w http.ResponseWriter, r *http.Request) {
	s.cacheHandler(w, r, "rss.xml")
}

func (s *Site) robotsTxtHandler(w http.ResponseWriter, r *http.Request) {
	s.cacheHandler(w, r, "robots.txt")
}

func (s *Site) faviconHandler(w http.ResponseWriter, r *http.Request) {
	s.cacheHandler(w, r, "static/favicon.ico")
}

func (s *Site) cacheHandler(w http.ResponseWriter, r *http.Request, key string) {
	// Try to get cached content
	item := s.cache.Get(key)
	if item == nil {
		s.handle404(w, r)
		return
	}

	if r.Header.Get("If-None-Match") == item.Etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", item.Mime)
	w.Header().Set("Cache-Control", item.CacheControl)
	w.Header().Set("Etag", item.Etag)
	w.WriteHeader(http.StatusOK)
	w.Write(*item.Content)
}

func (s *Site) handle404(w http.ResponseWriter, r *http.Request) {
	item := s.cache.Get("404")
	if item == nil {
		s.handle500(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusNotFound)
	w.Write(*item.Content)
}

func (s *Site) handle500(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)

	item := s.cache.Get("500")
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
