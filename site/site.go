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

type Site struct {
	port string
	env  string

	router *mux.Router

	pages     *PageManager
	posts     *PostManager
	assets    *AssetManager
	templates *template.Template
}

func NewSite(port string, env string) *Site {
	return &Site{
		port: port,
		env:  env,
	}
}

func (s *Site) Run() error {
	var err error

	// Load templates that we will use to render pages and posts
	s.templates, err = LoadTemplates(ContentDir)
	if err != nil {
		return err
	}

	s.posts = NewPostManager(PostsDir, s.templates)
	if err := s.posts.Load(); err != nil {
		return err
	}

	// Create caches for our various content types
	s.pages = NewPageManager(PagesDir, s.templates, s.posts)
	if err := s.pages.Load(); err != nil {
		return err
	}

	s.assets = NewAssetManager(AssetsDir)
	if err := s.assets.Load(); err != nil {
		return err
	}

	// Prepare routing
	router := mux.NewRouter()
	router.HandleFunc("/posts/{key}", s.postHandler).Methods("GET")
	router.HandleFunc("/static/{key}", s.staticHandler).Methods("GET")
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

func (s *Site) pageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	// The root page uses the "index" key
	if key == "" {
		key = "index"
	}

	// Try to get cache page
	page := s.pages.Get(key)
	if page == nil {
		s.Handle404(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	w.Write(*page.Content)
}

func (s *Site) postHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	// Try to get cache page
	post := s.posts.Get(key)
	if post == nil {
		s.Handle404(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	w.Write(*post.Content)
}

func (s *Site) staticHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	// Try to get cache page
	asset := s.assets.Get(key)
	if asset == nil {
		s.Handle404(w, r)
		return
	}

	log.Info(key, asset.Mime)
	w.Header().Set("Content-Type", asset.Mime)
	w.WriteHeader(http.StatusOK)
	w.Write(*asset.Content)
}

func (s *Site) Handle404(w http.ResponseWriter, r *http.Request) {
	page := s.pages.Get("400")
	if page == nil {
		s.Handle500(w, r)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html")
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

	w.Header().Set("Content-Type", "text/html")
	w.Write(*page.Content)
}
