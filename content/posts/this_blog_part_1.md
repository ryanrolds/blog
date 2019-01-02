# Blogging with Go, Markdown, and AWS 
<div id="created-at">2019-01-02T01:42:42Z</div>

This post goes over the architecture of the code used to serve this blog. The goal is to show just how much can be done with a little Go code, Markdown, and AWS. The post will show a basic HTTP service using Go's [`net/http`](https://golang.org/pkg/net/http/) package. How to use [Mux](https://github.com/gorilla/mux) to create endpoints for serving the home page, posts, and assets. Lastly, we will talk about using [Docker images](https://docs.docker.com/engine/reference/commandline/images/) and [AWS ECS](https://aws.amazon.com/ecs/) to make deployments a breeze.

> You don't need to know [Go](https://golang.org/) to understand this post. If you want to learn Go the best place to start is the [Go Tour](https://tour.golang.org). Once you're done with the tour, look at this project's `main.go` file and follow the code.

## Project structure

The source code can be found at [ryanrolds/pedantic_orderliness](https://github.com/ryanrolds/pedantic_orderliness). Despite there being about 3 dozen files, there isn't that much code. Two dozen of the files are assets and HTML templates. Only about 9 of the files actually contain Go code.

```
github.com/ryanrolds/pedantic_orderliness
├ content
│   ├ posts
│   │   ├ 2019_happy_new_year.md
│   │   └ blog_code.md
│   ├ static
│   │   ├ allow.txt
│   │   ├ bowman_lake_glacier_np.jpg
│   │   ├ disallow.txt
│   │   ├ email.png
│   │   ├ favicon.ico
│   │   ├ github.png
│   │   ├ instagram.png
│   │   ├ linkedin.png
│   │   ├ logo.png
│   │   ├ prog_intro_to_math.jpg
│   │   ├ ryanolds.jpg
│   │   ├ style.css
│   │   ├ thinking_fast_slow.jpg
│   │   └ twitter.png
│   ├ 404.md
│   ├ 500.md
│   ├ epilogue.tmpl 
│   ├ footer.tmpl
│   ├ header.tmpl
│   ├ index.tmpl
│   ├ layout.tmpl
│   ├ page.tmpl
│   ├ post.tmpl
│   ├ preamble.tmpl
│   └ sidebar.tmpl
├ site
│   ├ asset.go
│   ├ cache.go
│   ├ common.go
│   ├ page.go
│   ├ posts.go
│   ├ site.go
│   ├ static.go
│   └ templates.go
├ Dockerfile
├ Makefile
├ README.md
├ docker-compose.yml
├ main.go
└ pedantic_orderliness
```

> The above tree was made using the `tree` command. It's available on all major operating systems via a package manager (`apt-get install tree`, `yum install tree`, and `brew install tree`). If you're on Windows 10 you can install the Windows Subsystem for Linux, which will give you a decent version of Bash. This post is being written in `vim` on the WSL.

## Content directory

The `content` directory holds templates, some basic pages (Home, 404, 500), the `posts` directory (contains the Markdown files that will be made into posts), and finally, the `static` directory (contains all CSS, JS, images, and the robots.txt files). During service startup, all of these files are read into a Map (associative array, dictionary). Some of the files, templates (`.tmpl`) and Markdown (`.md`), are parsed and rendered. The rendered pages and posts are also stored in their maps/caches.

## Site directory

The `site` package is the primary package in the project. It initiates the loading of the `content` directory's files into their respective maps, sets up the HTTP endpoints and their handlers, and binds the HTTP server to port 8080 (or whatever port is provided by the `PORT` environment variable).

In the example below, the site logic (`site/site.go`) creates the endpoints that will be responding to HTTP requests. 

``` go
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
```
For example, the path of the page you're reading right now is `/posts/this_blog_part_1`.  Mux takes the path and tries to match against the defined endpoints. In this case it matches `/posts/{key}` with key being `this_blog_part_1` (above, line 3). I chose to use Mux because it has the same middleware and routing paradigm as Express.js.

Now that we know `s.postHandler` will be handling the request, let's look at it next.

``` go
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

  # Set http headers
  w.Header().Set("Cache-Control", "public, must-revalidate")
  w.Header().Set("Etag", post.Etag)
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  w.WriteHeader(http.StatusOK)

  # Write HTTP body
  w.Write(*post.Content)
}
```

All of the handlers (`postHandler`, `staticHandler`, ...) are pretty much the same. They get the `key` from the URL's path and use the key to look up the page/post/asset in the matching cache/map. The handler compares the looked up item's Etag (MD5 hash + file length) with the value of the `If-None-Match` header in the HTTP request to determine if browser needs to receive an updated version of the content/asset. Next, the handler sets some HTTP caching headers. 

The current version of the site uses the `must-revalidate` cache control directive, which isn't great. Each asset referenced by the HTML document still requires an HTTP request to confirm that a newer version isn't available. In the future, I will be switch `must-revalidate` out for `max-age=<seconds in one month>` and adding the MD5 hash to the query string of each asset referenced in the HTML document. This will probably be a future blog post as it's a pretty popular technique. MDN has a great page on [HTTP caching](https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching).

Once the headers have been written out, the actual HTTP body (CSS, images, HTML, ...) is written to the browser and the request is complete. That's pretty much it for the high-level stuff. There is a good chance that as I refine the code I will write additional posts about the internals.

## Running and Docker

Running the program is pretty straightforward. Check out the repo with Git and place it in your `$GOPATH`. Go to that directory and run `make install && make build && ./pedantic_orderliness` and then visit `http://localhost:8080` in your browser.

A `Dockerfile` and `docker-compose.yaml` files have also been created. If you have Docker installed and setup you can run the blog by simply running `docker-compose up`. The Dockerfile uses a multi-stage build and alpine images to keep the image size small, about 20MB. 

```
FROM golang:1.11.2-alpine3.8
RUN apk add --update make git
COPY . /go/src/github.com/ryanrolds/pedantic_orderliness
WORKDIR /go/src/github.com/ryanrolds/pedantic_orderliness
RUN make install
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY --from=0 /go/src/github.com/ryanrolds/pedantic_orderliness/pedantic_orderliness .
COPY --from=0 /go/src/github.com/ryanrolds/pedantic_orderliness/content content

CMD ["./pedantic_orderliness"]
```

## Hosting

I opted to use Amazon Web Services Elastic Container Service to host the site. I created an ECS Service that spreads 4 "tasks" (running instances of the blog) over 4 t3.nano EC2 instances (~$3.75/month each). The ECS Service creates a Target Group for the running instances. An Application Load Balancer can be pointed to that Target Group and configured to redirect HTTP to HTTPS (only shitty sites and neversll.com don't force HTTPS) and non-www to the www domain. 

The biggest value of this approach is that the site is highly available and deployments are a single line. Deploying a new version is as easy as running `make push_prod`. A series of commands are then run. First, a new Docker image is created and uploaded to AWS. Then the ECS Service is told creates 4 tasks/instances with the new image. When the new tasks are healthy (responding to health check requests) the older 4 tasks are drained, stopped, and destroyed. If new tasks fail, the old ones are left in place. Zero downtime. 

I also have a "test" ECS Service setup running on the same EC2 instances. As part of my development process, I deploy to "test" first and do a quick QA check before deploying to production.

## Wrap-up

I hope you find this post and the code samples useful. The primary take away is that it's possible to create fast responsive websites from pretty much scratch and host them in a highly available environment for not much time or money. It's my opinion that when it comes to website content, not much beats Markdown. It's concise, flexible, and doesn't have all the cruft that comes with more complicated formats. 

One final point, some of the patterns used in this project are also used in Node.js and Go web applications. The big difference being that web applications have database, dynamically render HTML documents, serve HTML that bootstrap Single Page Applications, and expose REST APIs.
