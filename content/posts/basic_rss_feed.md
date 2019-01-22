# RSS Feeds with Go's text/template
<div id="published-at">2019-01-22T00:16:19Z</div>

The blog is functional, but it’s missing features that improve reach and reader retention. Currently, readers must remember to visit the blog. That’s a tall order; We have to compete for their time and reading blogs is a low priority. Being able to display nudges where people spend their time is critical to winning some of their attention. Social media, newsletters, and RSS feeds are ways that we can fight for their attention.

In this post, we will look at a few different broadcasting techniques and implement an RSS feed. First, we will cover social media, newsletters, and RSS feeds. After narrowing in on RSS feeds, we will analyze an RSS feed's XML document. Using Go's `text/template` package we will create a template that renders an RSS feed for this blog. Lastly, we will look at the HTTP route handler that serves the document from `GET /rss.xml`.

## Social Media

Social media has become critical in reaching many demographics/markets. I've been posting to Facebook, Twitter, and LinkedIn, with decent success. Most of the blog's traffic comes from Twitter and Facebook, not direct links or search engines. With more posts and a little SEO work, hopefully, more traffic will be driven to the blog by Google, DuckDuckGo, Bing, and other search engines. In the short term, streamlining social media posts was the most impactful thing I could do.

A couple weeks ago, I added Facebook's Open Graph meta tags and Twitter's `card` meta tag:

``` html
<meta property="og:title" content="RSS Feeds with Go’s text/template">
<meta property="og:description" content="The blog is functional, but it’s missing features that improve reach and reader retention. Currently, readers must remember to visit the blog. That’s a tall order; We have to compete for their time and reading blogs is a low priority. Being able to display nudges where people spend their time is critical to winning some of their attention. Social media, newsletters, and RSS feeds are ways that we can fight for their attention.">
<meta property="og:url" content="https://test.pedanticorderliness.com/posts/basic_rss_feed">
<meta name="twitter:card" content="summary_large_image">
```

The above is rendered into this HTML document's `head` element using Go's `text/template` package. More on the Go's templating packages later. Rather than letting the social media providers guess at the key details of the post, we clearly define the values with meta tags. Now, when posting to Twitter, Facebook, or LinkedIn, the post will automatically contain the correct title, description, images (coming soon), and URI.

## Newsletters

Newsletters are another great broadcast option, but take much more time to implement and maintain. There are a few major pieces - the sign-up form, CAN-SPAM compliance, writing the newsletter, and debugging templates. The sign-up form and CAN-SPAM compliance are not the most difficult part of sending newsletters. Free and paid services exist that will take most work implementing the sign-up processes and CAN-SPAN's technical requirements off your hands. The bulk of the work is writting the newsletter, fighting with email templates, and debugging email client rendering issues. The template section of this article should help demystify templates a bit, which should be helpful when implementing email and other templates.

## RSS Feeds

After a reader adds the feed to their RSS reader (native/mobile/web app or browser plugin), new posts will automatically showup in their RSS reader. You don't have to do any additional work, like social media and newsletters. This technique has been around for a while, most of the RSS specifications are from the early 2000s. Despite RSS feed usage declining, it is still useful in reaching some demographics (IT workers and developers). If this blog were about something less technical, I would be more inclined to forgo RSS feeds.

Implementing an RSS feed requires serving an RSS compliant XML document containing details about the blog and its post's. In the next few sections will cover the format, template, and HTTP route handler required to serve that XML document.

### RSS Format

An RSS feed contains enough details for an RSS reader display aggregated feeds/posts in a UI. An RSS compliant XML document includes a `channel` and a list of posts, called `items`. 

Here is an example of an RSS feed with a single post:

``` xml
<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>

    <title>Pedantic Orderliness</title>
    <description>An assortment of technical posts, projects, game reviews, and random musings by Ryan Olds.</description>
    <link>https://www.pedanticorderliness.com/</link>
    <atom:link href="https://www.pedanticorderliness.com/rss.xml" rel="self" type="application/rss+xml" />
    <pubDate>Tue, 22 Jan 2019 01:10:35 +0000</pubDate>
    <ttl>1440</ttl>

    <image>
        <url>https://www.pedanticorderliness.com/static/logo.png</url>
        <title>Pedantic Orderliness</title>
        <link>https://www.pedanticorderliness.com/</link>
    </image>
    
    <item>
        <title>RSS Feeds with Go’s text/template</title>
        <description>The blog is functional, but it’s missing features that improve reach and reader retention. Currently, readers must remember to visit the blog. That’s a tall order; We have to compete for their time and reading blogs is a low priority. Being able to display nudges where people spend their time is critical to winning some of their attention. Social media, newsletters, and RSS feeds are ways that we can fight for their attention.</description>
        <link>https://test.pedanticorderliness.com/posts/basic_rss_feed</link>
        <guid>https://test.pedanticorderliness.com/posts/basic_rss_feed</guid>
        <pubDate>Tue, 22 Jan 2019 00:16:19 +0000</pubDate>
    </item>
    
    <item>
        <title>Efficient HTTP caching</title>
        <description>In this post, we will be talking about HTTP caching headers and strategies for maximizing browser-level caching while ensuring freshness. First, we will dive into HTTP and it’s caching headers. Then we will cover a few common strategies. And finally, we will review this blog’s implementation.</description>
        <link>https://test.pedanticorderliness.com/posts/efficient_http_caching</link>
        <guid>https://test.pedanticorderliness.com/posts/efficient_http_caching</guid>
        <pubDate>Fri, 04 Jan 2019 01:27:25 +0000</pubDate>
    </item>
    
    ...

</channel>
</rss>
```
> A quick note on reading XML and it's subset HTML. XML documents are nodes/elements (`<rss>`, `<channel>`, `<item>`, etc...) in a tree structure. Nodes can have attributes (`version`, `encoding`, `href`, etc...), which in turn have values. [Attribute-value pairs](https://en.wikipedia.org/wiki/Attribute%E2%80%93value_pair), also called key-value pairs, are used often by software developers. You will see the pattern used in JSON, YAML, CSS. Sets of key-value pairs, like attributes on nodes/elements, are a dictionary/map. Modern languages have parsing libraries that provide easy access to the document nodes/elements and their key-value pairs. As this is a tree, nodes can contain other nodes and a leaf/value.

Going through the above example line by line we see the XML declaration, `<?xml ...>`, and the root element/node, `<rss ...>`. The former is how the RSS reader (client) knows this is an XML document. After downloading and parsing the XML document, the client will query the tree for the root node. Then the client checks the root node for a `channel` node. With the `channel` node, the client retrieves the channel details and the list of items/posts. The channel's `ttl` and `pubDate` are used to determine how often to redownload and process the feed/document. A channel `image` is also defined, clients will usually display the image in their UI.

RSS readers will consume the channel's items/posts and display them in a list among posts from other feeds. The list item will show the post's `title`, `description`, `pubDate`, and `image` (if provided). See the [RSS 2.0 Specification](http://www.rssboard.org/rss-specification) for additional nodes/attributes that you can use in an RSS feed.

> The RSS format is flexible; You can use it for more than blog posts. I've seen it used to pass events between programs. Another way to think about it, RSS feeds allow an RSS reader (client) to consume posts (events) from a blog (service).

### Template

Armed with a decent idea about the structure of the XML document, we have a couple of options:

* Programmatically create the tree and the attributes of each node, then generate the XML document from the tree
* Define a template that renders a valid RSS XML document

This article is going to go with the 2nd option. The 1st option is perfectly valid and is preferred when the data is more complicated. Go does have an XML package that will get the job done, but it's more complicated than we need. Also, we can already easily load and render [Go templates](https://golang.org/pkg/text/template/) in this project. 

The template inserts the generation date (now) into the channel's `pubDate` and creates a channel `item` for each provided post:

``` xml
<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>

    <title>Pedantic Orderliness</title>
    <description>An assortment of technical posts, projects, game reviews, and random musings by Ryan Olds.</description>
    <link>https://www.pedanticorderliness.com/</link>
    <atom:link href="https://www.pedanticorderliness.com/rss.xml" rel="self" type="application/rss+xml" />
    <pubDate>{{ FormatRssDate .Generated}}</pubDate>
    <ttl>1440</ttl>

    <image>
        <url>https://www.pedanticorderliness.com/static/logo.png</url>
        <title>Pedantic Orderliness</title>
        <link>https://www.pedanticorderliness.com/</link>
    </image>

    {{ range .Posts}}
    <item>
        <title>{{ .Title }}</title>
        <description>{{ .Intro }}</description>
        <link>{{ .Url }}</link>
        <guid>{{ .Url }}</guid>
        <pubDate>{{ FormatRssDate .PublishedAt }}</pubDate>
    </item>
    {{ end }}

</channel>
</rss>
```

The template looks a lot like the XML document in the previous section. The major differences are the channel's `pubDate` element; The element contains a template directive that prints the document's generation date & time in a format required by the RSS specification. The next major difference is the `{{ range .Posts }}...{{ end }}` directive. The `range` directive iterates through the list and renders it's body (lines between `range` and `end` directives) once for each item in the list. Directives in the body are scoped to the list item. The body adds item elements for the `title`, `intro`, `link`, `guid` (we use the post's URI), and `pubDate`.

The code below renders the document using the "rss.tmpl" template file, generated at data+time, and a list of posts:

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

The `PageManager` struct has a method called `buildRss`. That method gets up to 20 posts from the `Postmanager` provided to the `PageManager`. It then creates a buffer that will hold the rendered XML document. The template is then rendered using the list of recent posts, the Site struct, and current date & time. Then the method adds the rendered document to the page cache under the key `rss.xml` (it's key+value store). With the page in the cache, we can now create an HTTP route handler that will respond with the cached document.

> When implementing the template, it's helpful to set up the HTTP route handler, run the server, and access the document with your browser. Copy the document text and run it through the [RSS Feed Validator](https://validator.w3.org/feed/#validate_by_input). 

### Route Handler

Let's get right to the handler code:

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

We've looked at similar code in a [previous post](/posts/this_blog_part_1). The handler determines the key to use when looking up a page in the cache. If it can't find a page, it returns a 404. Otherwise, it looks at an HTTP caching related header to determine if the client's copy is fresh (more on HTTP caching in a [previous post](/posts/efficient_http_caching). If the browser doesn't have a cached copy or the copy is old, write out the headers, 200 status code, and RSS document. 


Finally, we need to instruct the HTTP server to call this handler for specific HTTP paths:

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

We are using Go's [net/http package](https://golang.org/pkg/net/http/) and [Gorilla Mux](https://github.com/gorilla/mux) for routing. In an [earlier post](/posts/this_blog_part_1), we covered the HTTP server and router.

## Wrap-up

With a little background on the RSS document structure, a template made with Go's `text/template` package, and a Mux route handler, it's not a lot of code to add a [blog feed](/rss.xml). In the next post, we will go over logging to CloudWatch and some refactoring.
