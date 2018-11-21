package site

import(
  "fmt"
  "log"
  "net/http"

  "github.com/ryanrolds/pedantic_orderliness/page_cache"
)

type Site struct {
  Port string
  cache *page_cache.Cache
}

func (s *Site) Run() error {
  s.cache = page_cache.NewCache()

  server := http.Server{
    Addr: fmt.Sprintf(":%s", s.Port),
    Handler: s,
  }

  err := server.ListenAndServe()
  if err != nil {
    return err
  }

  return nil
}

func (s *Site) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
  log.Print(r.URL.Path)

  path := r.URL.Path
  if path == "/" {
    path = "/index"
  }

  log.Print(path)

  page := s.cache.Get(path)
  // If we couldn't find a page, then redirect to 404
  if page == nil {
    http.Redirect(rw, r, "/404", http.StatusTemporaryRedirect)
    return
  }

  rw.Header().Set("Content-Type", "text/html")
  page.Content.WriteTo(rw)
}
