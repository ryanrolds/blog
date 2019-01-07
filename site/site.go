package site

import (
	"fmt"
	"net/http"
	//"strings"
	"text/template"
	"time"

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

func (s *Site) Run() error {
	err := s.LoadContent()
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/static/{key}", s.AssetHandler)
	r.HandleFunc("/posts/{key}", s.PostHandler)
	r.HandleFunc("/robots.txt", s.RobotsTxtHandler)
	r.HandleFunc("/rss.xml", s.RssHandler)
	r.HandleFunc("/favicon.ico", s.RssHandler)
	r.HandleFunc("/{key}", s.PageHandler)
	r.HandleFunc("/", s.IndexHandler)
	r.HandleFunc("", s.IndexHandler)

	// Prepare server
	server := http.Server{
		Addr:         fmt.Sprintf(":%s", s.port),
		Handler:      r,
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

func (s *Site) LoadContent() error {
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

	return nil
}

func (s *Site) AssetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	s.CacheHandler(w, r, assetsDir+key)
}

func (s *Site) PostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	s.CacheHandler(w, r, postsDir+key)
}

func (s *Site) PageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	s.CacheHandler(w, r, postsDir+key)
}

func (s *Site) IndexHandler(w http.ResponseWriter, r *http.Request) {
	s.CacheHandler(w, r, "index")
}

func (s *Site) RssHandler(w http.ResponseWriter, r *http.Request) {
	s.CacheHandler(w, r, "rss.xml")
}

func (s *Site) FaviconHandler(w http.ResponseWriter, r *http.Request) {
	s.CacheHandler(w, r, "favicon.ico")
}

func (s *Site) RobotsTxtHandler(w http.ResponseWriter, r *http.Request) {
	s.CacheHandler(w, r, "robots.txt")
}

func (s *Site) CacheHandler(w http.ResponseWriter, r *http.Request, key string) {
	// Try to get cached content
	item := s.cache.Get(key)
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
