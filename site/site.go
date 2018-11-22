package site

import(
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "text/template"

  "github.com/ryanrolds/pedantic_orderliness/page_cache"
)

type Site struct {
  Port string
  template *template.Template
  cache *page_cache.Cache
}

func NewSite(port string) (*Site, error) {
  templateFile, err := ioutil.ReadFile("./site/template.html")
  if err != nil {
    log.Print("Cannot load ./site/template.html")
    return nil, err
  }

  template, err := template.New("home").Parse(string(templateFile[:]))
  if err != nil {
    log.Print("Unable to parse ./site/template.html")
    return nil, err
  }

  return &Site{
    Port: port,
    template: template,
    cache: page_cache.NewCache(),
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
