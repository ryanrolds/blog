package site

import(
  "fmt"
  "net/http"

  "github.com/ryanrolds/pedantic_orderliness/page_cache"
)

type Site struct {
  Port string
  cache *page_cache.Cache
}

func NewSite(port string) (*Site, error) {
  cache, err := page_cache.NewCache()
  if err != nil {
    return nil, err
  }

  return &Site{
    Port: port,
    cache: cache,
  }, nil
}

func (s *Site) Run() error {
  server := http.Server{
    Addr: fmt.Sprintf(":%s", s.Port),
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
  if path == "/" {
    path = "/index"
  }

  page := s.cache.Get(path)
  // If we couldn't find a page, then redirect to 404
  if page == nil {
    page = s.cache.Get("/404")
    rw.WriteHeader(http.StatusNotFound)
  } else{
    rw.WriteHeader(http.StatusOK)
  }

  rw.Header().Set("Content-Type", "text/html")
  rw.Write(*page.Content)
}
