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
	ContentDir  = "./content/"
	TemplateDir = ContentDir
	PagesDir    = ContentDir
	PostsDir    = ContentDir + "posts/"
	AssetsDir   = ContentDir + "static/"
)

type Hashes map[string]string

type Site struct {
	port   string
	Env    string
	Hashes *Hashes

	router    *mux.Router
	cache     *ContentCache
	templates *template.Template
}

func NewSite(port string, env string) *Site {
	return &Site{
		port: port,
		Env:  env,
	}
}

func (s *Site) Run() error {
	cache := NewContentCache()

	err := loadAssets(AssetsDir, cache)

	template, err := loadTemplates(ContentDir)
	if err != nil {
		return err
	}

	posts, err = loadPosts(PostsDir)
	if err != nil {
		return err
	}

	err = loadPages(PagesDir, templates, posts)
	if err != nil {
		return err
	}

	s.Hashes = s.assets.GetHashes()

	// Prepare routing
	router := mux.NewRouter()
	router.HandleFunc("/posts/{key}", s.postHandler).Methods("GET")
	router.HandleFunc("/static/{key}", s.staticHandler).Methods("GET")
	router.HandleFunc("/favicon.ico", s.faviconHandler).Methods("GET")
	router.HandleFunc("/robots.txt", s.robotsHandler).Methods("GET")
	router.HandleFunc("/{key}", s.pageHandler).Methods("GET")
	router.HandleFunc("/", s.pageHandler).Methods("GET")
	router.HandleFunc("", s.pageHandler).Methods("GET")
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
	if content == nil {
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
	page := s.pages.Get("404")
	if page == nil {
		s.Handle500(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write(*page.Content)
}

func (s *Site) Handle500(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)

	page := s.pages.Get("500")
	if page == nil {
		log.Warn("Unable to get 500 page")
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(*page.Content)
}

func getPostUrl(env string) string {
	domain := "localhost:8080"
	if env == "production" {
		domain = "www.pedanticorderliness.com"
	} else if env == "test" {
		domain = "test.pedanticorderliness.com"
	}

	return domain
}
