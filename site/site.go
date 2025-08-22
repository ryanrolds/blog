package site

import (
	"embed"
	"fmt"
	"net/http"
	"strconv"

	//"strings"
	"text/template"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	ContentDir  = "./content/"
	TemplateDir = ContentDir
	PagesDir    = ContentDir
	PostsDir    = ContentDir + "posts/"
	AssetsDir   = ContentDir + "static/"
)

type Hashes map[string]string

var ContentFS embed.FS

type Site struct {
	port   string
	Env    string
	Log    *logrus.Entry
	Hashes *Hashes

	router *mux.Router

	pages     *PageManager
	posts     *PostManager
	assets    *AssetManager
	templates *template.Template
}

func NewSite(port string, env string, log *logrus.Entry) *Site {
	return &Site{
		port: port,
		Env:  env,
		Log:  log,
	}
}

func (s *Site) SetContentFS(fs embed.FS) {
	ContentFS = fs
}

func (s *Site) Run() error {
	var err error

	s.assets = NewAssetManager("")
	if err := s.assets.Load(); err != nil {
		return err
	}

	s.Hashes = s.assets.GetHashes()

	// Load templates that we will use to render pages and posts
	s.templates, err = LoadTemplates("content")
	if err != nil {
		return err
	}

	s.posts = NewPostManager(s, "", s.templates)
	if err := s.posts.Load(); err != nil {
		return err
	}

	// Create caches for our various content types
	s.pages = NewPageManager(s, "", s.templates, s.posts)
	if err := s.pages.Load(); err != nil {
		return err
	}

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

	loggingHandler := handlers.LoggingHandler(s.Log.Writer(), router)

	// Prepare server
	server := http.Server{
		Addr:         fmt.Sprintf(":%s", s.port),
		Handler:      loggingHandler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	s.Log.Infof("Starting server on port %s", s.port)

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
		
		// Handle pagination for index page
		pageParam := r.URL.Query().Get("page")
		if pageParam != "" {
			pageNum, err := strconv.Atoi(pageParam)
			if err != nil || pageNum < 1 {
				pageNum = 1
			}
			
			page := s.pages.GetPaginated("index", pageNum)
			if page == nil {
				s.Handle404(w, r)
				return
			}
			
			if r.Header.Get("If-None-Match") == page.Etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			w.Header().Set("Content-Type", page.Mime)
			w.Header().Set("Cache-Control", page.CacheControl)
			w.Header().Set("Etag", page.Etag)
			w.WriteHeader(http.StatusOK)
			w.Write(*page.Content)
			return
		}
	}

	// Try to get cache page
	page := s.pages.Get(key)
	if page == nil {
		s.Handle404(w, r)
		return
	}

	if r.Header.Get("If-None-Match") == page.Etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", page.Mime)
	w.Header().Set("Cache-Control", page.CacheControl)
	w.Header().Set("Etag", page.Etag)
	w.WriteHeader(http.StatusOK)
	w.Write(*page.Content)
}

func (s *Site) postHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	// Try to get cache entry for post
	post := s.posts.Get(key)
	if post == nil {
		s.Handle404(w, r)
		return
	}

	if r.Header.Get("If-None-Match") == post.Etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, must-revalidate")
	w.Header().Set("Etag", post.Etag)
	w.WriteHeader(http.StatusOK)
	w.Write(*post.Content)
}

func (s *Site) staticHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	// Try to get cached entry for asset
	asset := s.assets.Get(key)
	if asset == nil {
		s.Handle404(w, r)
		return
	}

	if r.Header.Get("If-None-Match") == asset.Etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", asset.Mime)
	w.Header().Set("Cache-Control", "public, max-age=2419200")
	w.Header().Set("Etag", asset.Etag)
	w.WriteHeader(http.StatusOK)
	w.Write(*asset.Content)
}

func (s *Site) faviconHandler(w http.ResponseWriter, r *http.Request) {
	asset := s.assets.Get("favicon.ico")
	if asset == nil {
		s.Handle404(w, r)
		return
	}

	w.Header().Set("Content-Type", asset.Mime)
	w.Header().Set("Cache-Control", "public, max-age=604800")
	w.Header().Set("Etag", asset.Etag)
	w.WriteHeader(http.StatusOK)
	w.Write(*asset.Content)
}

func (s *Site) robotsHandler(w http.ResponseWriter, r *http.Request) {
	robotsFile := "allow.txt"
	if s.Env != "production" {
		robotsFile = "disallow.txt"
	}

	asset := s.assets.Get(robotsFile)
	if asset == nil {
		s.Handle404(w, r)
		return
	}

	w.Header().Set("Content-Type", asset.Mime)
	w.WriteHeader(http.StatusOK)
	w.Write(*asset.Content)
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
	page := s.pages.Get("500")
	if page == nil {
		s.Log.Warn("Unable to get 500 page")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(*page.Content)
}
