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
	env     string
	rootDir string

	router    *mux.Router
	cache     *Cache
	templates *template.Template
	posts     *PostList
	hashes    *Hashes
}

func NewSite(port, env, contentDir string) *Site {
	return &Site{
		port:    port,
		env:     env,
		rootDir: contentDir,
	}
}

func (s *Site) Run() error {
	s.cache = NewCache()

	err := LoadAssets(s.rootDir+assetsDir, s.cache)

	s.templates, err = LoadTemplates(s.rootDir)
	if err != nil {
		return err
	}

	s.posts, err = LoadPosts(s, s.rootDir+postsDir, s.templates, s.cache)
	if err != nil {
		return err
	}

	err = LoadPages(s, s.rootDir, s.templates, s.posts, s.cache)
	if err != nil {
		return err
	}

	s.hashes = s.cache.GetHashes()

	// Prepare routing
	router := mux.NewRouter()
	router.HandleFunc("/{key}", s.contentHandler).Methods("GET")
	router.HandleFunc("/", s.contentHandler).Methods("GET")
	router.HandleFunc("", s.contentHandler).Methods("GET")
	s.router = router

	// Prepare server
	server := http.Server{
		Addr:         fmt.Sprintf(":%s", s.port),
		Handler:      s.router,
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

func (s *Site) contentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	// The root page uses the "index" key
	if key == "" {
		key = "index"
	}

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
	item := s.cache.Get("404")
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
