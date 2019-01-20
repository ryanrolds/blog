# RSS Feeds with Go's text/template

In this post we will be covering RSS 2.0 Feeds and a little bit of Go's [`text/template`](https://golang.org/pkg/text/template/) package. We will look at RSS's format and it's relationship to XML. Implement a template that produces and XML document conforming to the RSS 2.0 specification. Finally, we will implement an HTTP route handler that will response with the blogs rendered RSS XML document..

### Why?

An RSS Feed is an important blog feature, it provides a way to broadcast new posts to your readers/users. Without RSS Feeds, and other methods of broadcasting, your readers would have to remember to visit your site to discovery new content. Broadcasting new posts to your users will drive views and improve reader retention. 

A newsletter is another broacast option, newsletters are much more effective then an RSS Feed, but it's also much more work. A newletter sign-up form isn't the most difficult part of newsletters. Free and paid services exists that will take most of the sign-up processes technical burden off your hands. The bulk of the work is actually sending the newsletter, fighting with email templates, and debugging email client rendering issues. The template section of this article should help demystify templates, which may be helpful when implementing email, and other, templates. 

## Format

RSS 2.0 Feeds are XML Documents that contain a `channel` and a list of posts, called `items`. An RSS Feed contains enough details about a blog and it's posts that an RSS Reader can aggregate and display the feeds/posts in a UI. Here is an example of an RSS feed with a single post:

``` xml
<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
    <title>Pedantic Orderliness</title>
    <description>An assortment of technical posts, projects, game reviews, and
        random musings by Ryan Olds.</description>
    <link>https://www.pedanticorderliness.com/</link>
    <image>
        <url>https://www.pedanticorderliness.com/static/logo.png</url>
        <title>Pedantic Orderliness</title>
        <link>https://www.pedanticorderliness.com/</link>
    </image>
    <pubDate>Sat, 19 Jan 2019 00:21:57 +0000</pubDate>
    <ttl>1440</ttl>
    <atom:link href="https://www.pedanticorderliness.com/rss.xml"
        rel="self" type="application/rss+xml" />
    
    <item>
        <title>Efficient HTTP caching</title>
        <description>In this post, we will be talking about HTTP caching headers
            and strategies for maximizing browser-level caching while ensuring
            freshness. First, we will dive into HTTP and it’s caching headers.
            Then we will cover a few common strategies. And finally, we will review
            this blog’s implementation..</description>
        <link>https://www.pedanticorderliness.com/posts/efficient_http_caching</link>
        <guid>https://www.pedanticorderliness.com/posts/efficient_http_caching</guid>
        <pubDate>Fri, 04 Jan 2019 01:27:25 +0000</pubDate>
    </item>
    
</channel>
</rss>
```
> Quick note on reading XML and it's subset HTML. XML documents are nodes/elements (`<rss>`, `<channel>`, `<item>`, etc...) in a tree structure. Nodes can have attributes (`version`, `encoding`, `href`, etc...), which in turn have values. [Attribute-value pairs](https://en.wikipedia.org/wiki/Attribute%E2%80%93value_pair), also called a key-value pairs, are a _very_ common pattern in software. You will see the pattern used in JSON, YAML, CSS. Think of key-value pairs as a dictionary/map containing details of the node. XML parsing libraries provide easy access to nodes and their key-value pairs. As this is a tree, nodes can contain other nodes, and/or a leaf/text.

Going through the above example line by line we see the XML deleration, `<?xml ...>`, and the root element/node, `<rss ...>`. The former is how the RSS Reader (client) knows this is an XML document. The root element is the top most node in the document's tree. After the client downloads the XML document it parses it with a library, like [XPath](https://en.wikipedia.org/wiki/XPath). The library will provide easy to use functions to traverse and/or query the document tree. 

The client retieves the channel and it's list of items/posts. The channel's `ttl` and `pubDate` are used to determine how often to redownload and process the feed/document. A channel image is also defined, clients will often display the channel image in their UI. The purpose title, description, link should be obvious.  

RSS Readers will consume the channel's items/posts and display them in a list among posts from other feeds. Normally the lists are sorted by the post's `pubDate`. The list item will display the post's title, description, pubDate, and image (if provided). There are many other nodes that can be defined in an RSS XML document. See the [RSS 2.0 Specification](http://www.rssboard.org/rss-specification) and please don't expect specification pages to be pretty.

> The RSS format is flexible, it doesn't just have to be used for blog posts. I've seen it used to pass basic events between programs. Another way to think about it, RSS feeds allow an RSS Reader (client) to consume posts (events) from a blog (service).

## Template

Armed with a decent idea about the structure of the XML document that we must create we have a couple of options:

* Progromatically create the tree and define each nodes attributes, then generate the XML document from the tree
* Define a template that renders a valid RSS XML document

This article is going to go with the 2nd option. The 1st option is perfectly valid and is prefered when the data the document is generated from is more complex. Go does have an XML package that will get the job done, but it's a much more complex solution then we need. Also, we are already able to easily load and render Go templates in this project. 

It's worth restating our purpose at this point, we are creating an RSS Feed for our site. The feed will contain some basic information about the blog and a list of recent posts. The template will be off a basic RSS feed with placeholders for the blog's details, we also need to iterate of a list of posts and render a channel item for each post.  template has been created at ./content/rss.tmpl:

``` xml
<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
    <title>Pedantic Orderliness</title>
    <description>An assortment of technical posts, projects, game reviews, and random musings by Ryan Olds.</description>
    <link>https://www.pedanticorderliness.com/</link>
    <image>
        <url>https://www.pedanticorderliness.com/static/logo.png</url>
        <title>Pedantic Orderliness</title>
        <link>https://www.pedanticorderliness.com/</link>
    </image>
    <pubDate>{{ FormatRssDate .Generated}}</pubDate>
    <ttl>1440</ttl>
    <atom:link href="https://www.pedanticorderliness.com/rss.xml" rel="self" type="application/rss+xml" />
    {{ range .Posts}}
    <item>
        <title>{{ .Title }}</title>
        <description>{{ .Intro }}.</description>
        <link>{{ .Url }}</link>
        <guid>{{ .Url }}</guid>
        <pubDate>{{ FormatRssDate .PublishedAt }}</pubDate>
    </item>
    {{ end }}
</channel>
</rss>
```

The template looks a lot like the XML document in the previous section. The major differences are the channel's pubDate element; The element's contents is a tempalte directive that prints the document's generation date & time in a format required by the RSS specification. The next major different is the `{{ range .Posts }}...{{ end }}` directive. When the tmeplate is rendered it's provided a list of posts, the `range` directive iterates through the list and any template directives before the `end` directive are scoped to an post in the provided list. The title, intro, link, guid (we use the post's url), and published date are for each post is rendered into a channel item.

Now we can look at the code that renders the document using the "rss.tmpl" template file and a list of posts:

``` go
const rssLimit = 20
const rssKey = "rss.xml"

func (p *PageManager) buildRss() error {
  // Get a list of most recent posts
  posts := p.posts.GetRecent(rssLimit)

  buf := &bytes.Buffer{}
  err := p.templates.ExecuteTemplate(buf, "rss.tmpl", &TemplateData{
    Title:      "",
    CSS:        "",
    JavaScript: "",
    Content:    "",
    Posts:      &posts,
    Social:     &Social{},
    Site:       p.site,
    Generated:  time.Now(),
  })
  if err != nil {
          return err
  }

  p.cache.Set(rssKey, &Page{
    Content:      &buf.Bytes(),
    Etag:         getEtag(&body),
    Mime:         "application/rss+xml; charset=utf-8",
    CacheControl: "public, must-revalidate",
  })

  return nil
}
```

The `PageManager` struct has a method called `buildRss`. That method gets up to 20 posts from the `Postmanager` provided to the `PageManager`. It then creates a buffer that will hold the rendered XML document. The template is then rendered using the list of recent posts, the Site struct, and current date & time. Then the method adds the rendered document to the page cache under the key `rss.xml` (it's key+value store). With the page in the cache we can now create an HTTP router handler that will respond with cached document.

> When implementing the template it's helpful to setup the HTTP route handler, run the server, and access the document with your browser. Copy the document text and run it through the [RSS Feed Validator](https://validator.w3.org/feed/#validate_by_input). 

## Route Handler

Lets get right to the handler code:

``` go
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
```

We've looked at similar code in a [pevious post](/posts/this_blog_part_1). The handler determines the key to use when looking up a page in the cache. If it can't find a page it returns a 404. Otherwise, it looks at an HTTP cacheing related header to determine if the client's copy if fresh (more on HTTP caching in a [previous post](/posts/efficient_http_caching). If the browser doesn't have a cached copy or the copy is old, write out the headers, 200 status code, and RSS document. 


Finally we need to instruct the HTTP server to call this handler for specific HTTP paths:

``` go
func (s *Site) Run() error {
  // Setup router
  ...
  router.HandleFunc("/{key}", s.pageHandler).Methods("GET")
  router.HandleFunc("/", s.pageHandler).Methods("GET")
  router.HandleFunc("", s.pageHandler).Methods("GET")
  ...
  // Run HTTP server with router as handler
}
```

We are using Go's [net/http package]() and [Gorilla Mux]() for routing. In an [earlier post]() we covered the http server and router. And now the blog has an [RSS Feed](/rss.xml).

## Wrap-up

That about does it for RSS feeds. With a little background on the RSS document structure, a template made with Go's `text/template` package, and Mux route handler it's not a lot of code to add an RSS feed. In the next post we will go over logging to CloudWatch and some refactoring.
