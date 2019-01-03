# Efficient HTTP cachine

In this post we will be talking about HTTP caching headers and strategies for maximizing browser-level caching while ensuring freshness. First we will drive into HTTP and it's caching headers. Then we will talk about a few common strategies. And finally, we will cover this blogs implementation.

## HTTP & Caching

This section is a quick review of HTTP and the common caching related headers. If you're already familure with the anatomy of an HTTP request and response, `Cache-Control`, `If-None-Modfied`, etc... then feel free to skip to the next section.

### Request and response

When you loaded this page your browser opened a connection/socket to www.pedanticorderilness.com on port 443. The HTTP/1.1 protocol dictates that a request be initiated with a Request-Line, Request Header, and Reqeust Body and that the response contain a Status-Line, Response Headers, and Response Body. The request for this page looks like:

```
GET /posts/efficient_http_caching HTTP/1.1
host: www.pedanticorderliness.com
accept: text/html
accept-encoding: gzip, deflate, br
accept-language: en-US,en;q=0.9,da;q=0.8
cache-control: max-age=0
dnt: 1
pragma: no-cache
referer: https://www.pedanticorderliness.com/
user-agent: example from blog post 
```

> The request and response data is viewable in your browser's developer tools. Also, if you were to use `openssl s_client`, you could establish a connection to this server and paste the above code, press enter a couple times, then you would get this page in response. I will leave the details as a search exercise for the curious readers.

The service hosting this states responds with:

```
HTTP/1.1 200 OK
Date: Thu, 03 Jan 2019 18:05:13 GMT
Content-Type: text/html; charset=utf-8
Transfer-Encoding: chunked
Connection: keep-alive
Cache-Control: public, must-revalidate
Etag: e66d665cd7a6d67ca6112b21c6351c1c_3317

<Response Body>
```

Why am I talking about this? It's to point out the Request and Response Headers. Everything to do with HTTP caching is about what headers the browsers sends and receive. 

### Headers

The request headers include `cache-control: max-age=0` and `pragma: no-cache` (legacy). The value `max-age=0` tells the server (and any intermediate caches, CDN) to not use a cached copy and to check the origin (the blog server). The `pragma` header is a legacy HTTP/1.0 header and is replaced by HTTP/1.1's `cache-control`, it's still useful to set as not every keeps their services up-to-date (but seriously we've been on HTTP/1.1 for nearly two decades, update your shit people).

The response headers contain `Cache-Control: public, must-revalidate`, `Date, Thu, 03 Jan 2019 18:05:13 GMT`, and `Etag: e66d665cd7a6d67ca6112b21c6351c1c_3317`. These headers instruct the browser (and intermediate caches, like CloudFlare) how to handle freshness. An HTTP responses `Cache-Control` header have a few directives. The value can contain `public` (should cache) or `private` (should not cache), it can also define an `max-age=<seconds>` or have `must-revalidate`. There are acutally quite a view other directives, check out MDN's [documenation for `Cache-Control`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control). 

The `ETag` header is part of set of headers that allow a server (or intermedate cache) to avoid resending the Response Body, which us usually much larger in size then the headers and cuts response times down by a lot (very desirable). If a browser has the response in it's cache already it can look up the `Etag` value and include `If-None-Match: <Etag value>` in the request headers. The server (or intermediate cache) can checks the ETag value it has for it's copy and if they match respond with a 304 Status Code and not send the Response Body. A similar mechanism exists with the `Lst-Modified` and `If-Modified_Since` (or `If-Unmodified-Since`) headers. I don't like using those headers because they only have a 1 second granularity. I prefer using an `Etag` with the value being the Response Body's MD5 sum and length.

## Caching Strategies

## This blog's implementation

## Wrap-up
