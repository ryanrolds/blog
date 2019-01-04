# Efficient HTTP caching
<div id="created-at">2019-01-05T01:27:25Z</div>

In this post, we will be talking about HTTP caching headers and strategies for maximizing browser-level caching while ensuring freshness. First, we will drive into HTTP and it's caching headers. Then we will talk about a few common strategies. And finally, we will cover this blog's implementation.

## HTTP & Caching

This section is a quick review of HTTP and the common caching-related headers. If you're already familiar with the anatomy of an HTTP request and response, `Cache-Control`, `If-None-Modified`, etc... then feel free to skip to the next section.

### Request and Response

When you loaded this page your browser opened a connection/socket to the blog. The HTTP/1.1 protocol dictates that a request is performed by sending a Request-Line, Request Header, and Request Body. The server looks at the request and responds with a Status-Line, Response Header, and Response Body. The request for this page looks like:

```
GET /posts/efficient_http_caching HTTP/1.1
Host: www.pedanticorderliness.com
Accept: text/html
Accept-Encoding: gzip, deflate, br
Accept-Language: en-US,en;q=0.9,da;q=0.8
Cache-Control: max-age=0
Pragma: no-cache
Referer: https://www.pedanticorderliness.com/
User-Agent: example from blog post 
```

> The request and response data are viewable in your browser's developer tools. Also, if you were to use `openssl s_client`, you could establish a connection to this server and paste the above code. After pressing enter a couple times, the server would respond with this page. I will leave the specific command to run as a search exercise for the more curious readers.

The server hosting this site responds with:

```
HTTP/1.1 200 OK
Date: Thu, 03 Jan 2019 18:05:13 GMT
Content-Type: text/html; charset=utf-8
Transfer-Encoding: chunked
Connection: keep-alive
Cache-Control: public, must-revalidate
ETag: e66d665cd7a6d67ca6112b21c6351c1c

<Response Body> ...
```

Why am I writing about this? It's to point out the Request and Response Headers. Everything to do with HTTP caching is about what headers the browser sends and receives.

### Headers

The request headers include `Cache-Control: max-age=0` and `Pragma: no-cache` (legacy). The directive  `max-age=0` tells the server (and any intermediate caches, like a CND or CloudFlare) to not use a cached copy and to check the origin (the blog server). `Pragma` is a legacy HTTP/1.0 header and is replaced by HTTP/1.1's `Cache-Control`, it's still useful to set as not everyone keeps their services up-to-date (but seriously we've had HTTP/1.1 for nearly two decades, WTF?).

The response headers contain `Cache-Control: public, must-revalidate`, `Date: Thu, 03 Jan 2019 18:05:13 GMT`, and `ETag: e66d665cd7a6d67ca6112b21c6351c1c`. These headers give the browser (and intermediate caches) the rules around caching this URI. An HTTP response's `Cache-Control` header has a few directives. The most common are `public` (should cache), `private` (should not cache), `max-age=<seconds>`, and `must-revalidate`. There are quite a view other directives, check out MDN's [documentation for `Cache-Control`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control). 

Keep in mind that `max-age=<seconds>` can have a different meaning depending on the context. If it's in the Request Headers it limits how old of a cache entry intermediate caches are allowed respond with. If the intermediate cache's entry is recent enough then the request never makes it to the origin server. If it's too old then the intermediate cache checks with the origin, possibly pulling a fresher response. If the directive in the Response Headers, it defines a period in which the asset an can simply be served from the browser's cache, no request needed. Once this period passes then it must be refreshed/revalidated.

The `ETag` header is part of a set of headers that allow a server (or intermediate cache) to avoid unnecessarily sending the Response Body, which is usually much larger in size than the Response Headers. Not sending the body reduces response times by a lot (very desirable, especially on mobile). If a browser has the response in its cache it can look up the `Etag` value and include `If-None-Match: <Etag value>` in the Request Headers. The server (or intermediate cache) can check the `ETag` value it has and if they match, respond with a `304 Not Modified` status code and not send the Response Body.

A similar mechanism exists with the `Last-Modified` and `If-Modified-Since` (or `If-Unmodified-Since`) headers. I don't like using those headers because they only have a 1-second granularity, a lot can happen in one second. I prefer using an `Etag` with the value being the Response Body's MD5 sum.

## Caching Strategies

With all of these directives, we have quite a bit of flexibility. Our primary goal is to provide the fastest and most responsive experience to our users. To do this we don't want to cache volatile things, or at confirm it didn't change and not perform requests when we know the response can't change. Ultimately there only a few buckets that URIs and their responses can fall into:

* Immutable - In a perfect world we only ever serve this response once per browser/cache.
* Mostly static - Doesn't change, prime for caching
* Volatile - Changes with every request, trickier to cache
* Sensitive - Contains data that shouldn't be cached
* Unsafe - PUT, POST, DELETE requests

> I rarely see the `immutable` directive used, it's just too easy to shoot yourself in the foot and having to burn the URI (never use it again because you can never know for sure that someone doesn't have an old version cached).

We have a lot of options available and one strategy doesn't fit all situations. When deciding if something should be cached and for how long you, ask yourself a few questions:

* Is the response sensitive?
* Is it a safe request? GET, HEAD, OPTIONS
* Does the response change often?
* How often are users going to request this URI? (referenced in HTML, CSS, JS)
* Is the response an HTML document that may reference other assets that may have changed? 

Things that are not volatile and are used often deserve aggressive caching. Things that change often shouldn't be cached or at least should be revalidated often. Sensitive data should not be cached (`private` directive), period. `PUT`, `POST`, and `DELETE` requests are not safe to cache.

One problem that comes up is needing to bust the cache. This problem usually manifests as someone needed to deploy a bug fix quickly and they don't want anyone to continue using the old/buggy version. Your only option is to deploy code that uses a new URI for the affected asset. Some projects put versions in the filename/path and others append a version or hash to the query string. I usually go with the hash route, because who wants to version every asset? The code for storing and managing a version for every asset is non-trivial. Using some kind of hashing (MD5, SHA, etc...) avoids having to store asset versions because the hash is intrinsic to the asset's data, changing the asset is changing the MD5.

A less common problem is when you are varying the response based on the `User-Agent` header, or some other header. In those cases, you must properly set the `Vary` header so that intermediate caches don't serve the wrong content.

## The Blog's Implementation

First, we need to define the strategy we will use. All responses will have an `ETag` header that contains the MD5 sum of the Response Body. Responses containing HTML will have the `must-revalidate` directive, that way we can be sure users always have a fresh version but we can also avoid having to resend what they already have.

Assets referenced in the HTML will have the MD5 sum in the query string and max age of 1 month. During startup, we will read all of our assets from `./content/static` and generate the MD5 sums. We will make a `map[string]string` that maps the key to the MD5 sum. Lastly, we will implement a template helper function that will take the key, look up the MD5 sum, and return the complete URL for the asset. As long as we remember to use the helper function we shouldn't have to think about caching anymore. Assets that change will have a new MD5 hash and URL, assets that stay the same will have the same MD5 hash and URL.

No point in littering the code with calls to Go's `md5.Sum(...)`, so lets put a helper function it in our "common" file.

``` go
import "crypto/md5"
func getEtag(buffer *[]byte) string {
	hash := md5.Sum(*buffer)
	return fmt.Sprintf("%x", hash)
}
```

Every time we read a file and add it the page/post/asset cache, we will call the above function to generate an MD5 hash. Below is an example of the asset processing logic.

``` go
func (p *AssetManager) buildAsset(filename string) (*Asset, error) {
    // Get byte array and mime type (text/css, image/png, etc...)
	buffer, mime, err := getAsset(filename)
	if err != nil {
		return nil, err
	}

	return &Asset{
		Mime:    mime,
		Content: buffer,
		Etag:    getEtag(buffer),
	}, nil
}
```

Now that every post, page, and asset have an MD5 sum in it's struct, we need to build a map that we can use in our templates. We don't need the MD5 sum for posts or pages in our map, but we do need assets (images, CSS, JS) in the map. 

``` go
type Hashes map[string]string
func (p *AssetManager) GetHashes() *Hashes {
	hashes := Hashes{}

	keys := p.cache.GetKeys()
	for _, key := range keys {
		value := p.cache.Get(key)
		hashes[key] = value.(Asset).Etag
	}

	return &hashes
}
```

It would be annoying to have to add `?m={{.Site.Hashes[assetKey]}}` all over the templates, so we will use a helper function.

``` go
tmpl := template.New("").Funcs(template.FuncMap{
    "GetAssetURL": func(key string, hashes *Hashes) string {
        return fmt.Sprintf("/static/%s?m=%s", key, (*hashes)[key])
    },
})

// Template example
// <img src="{{ GetAssetURL "logo.png" .Site.Hashes }}"/>
```

We will still need the MD5 sum for posts and pages so that we can handle the `If-None-Match` header.

``` go
if r.Header.Get("If-None-Match") == post.Etag {
    w.WriteHeader(http.StatusNotModified)
    return
}

w.Header().Set("Content-Type", "text/html; charset=utf-8")
w.Header().Set("Cache-Control", "public, must-revalidate")
w.Header().Set("Etag", post.Etag)
w.WriteHeader(http.StatusOK)
w.Write(*post.Content)
```

## Wrap-up

That's it, Efficient HTTP caching. My approach isn't the only approach. Different situations are going to require different approaches to caching. One final very important point. No matter how you implement caching, not just HTTP caching, know how you will evict/bust your cache. You do not want to be in a place where you need to bust the cache and cant. This is important, never implementing a cache without also implementing a way to bust it.
