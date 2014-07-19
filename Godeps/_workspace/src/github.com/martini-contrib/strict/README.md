# [Martini][1] Strict Mode [![wercker status](https://app.wercker.com/status/3adeb16c006087c9a999da4084288241/s/ "wercker status")](https://app.wercker.com/project/bykey/3adeb16c006087c9a999da4084288241)


[1]: //github.com/go-martini/martini

This repo contains a set of utilities that help you make a well-behaving,
strict API using the awesome Martini framework. The are tested and ready-to-use
handlers for the following responses:

* 404 Not Found with empty body
* 405 Method Not Allowed + Allow header
* 406 Not Acceptable
* 415 Unsupported Media Type

There is also a helper function to negotiate the request content type,
according to [RFC2616 Section 14][4].


## Usage

Here is a complete working example that uses the `strict` package together with
the [`render`][2] contrib package. In particular, the `render` and [`binding`][3]
contrib packages work very nicely together with `strict`.

[2]: https://github.com/martini-contrib/render
[3]: https://github.com/martini-contrib/binding

```go
package main

import (
	"net/http"

	"github.com/attilaolah/strict"
	"github.com/go-martini/martini"
)

func main() {
	m := martini.Classic()
	m.Use(strict.Strict)
	m.Use(render.Renderer())

	m.Get("/zoo", strict.Accept("application/json", "text/html"), func(n strict.Negotiator) {
		// This will only run if the Accept header was either empty or included
		// application/json, application/*, text/html, text/* or */*.
		// n.Accepts("application/json") can be used to check which content type is preferred.
		if n.Accepts("application/json") > n.Accepts("text/html") {
			// JSON is preferred, return encoded output.
		}
		// HTML is preferred (or both content types have an equal q value), render template.
	})
	m.Post("/zoo", strict.ContentType("application/json", "text/xml", ""), func(n strict.Negotiator) {
		// This will only run if the content-type header was either application/json, text/xml or empty.
		// n.ContentType("text/xml") can be used to checx if the content type was xml.
	})

	// 405 for PUT, PATCH, DELETE, etc.
	m.Router.NotFound(strict.MethodNotAllowed, strict.NotFound)

	m.Run()
}
```


#### The `strict.Strict` handler

By telling Martini to `m.Use(strict.Strict)`, the `strict.Negotiator` interface
becomes availabe in handlers. The negotiator can be used for two things:

* It can check whether the client accepts a content type by parsing the
  `Accept` header (if present) and returning the corresponding `q` value. See
  [RFC2616 Section 14][4] for details.

[4]: http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html

Calling `n.Accepts("text/html")` will return the client's preference to accept
the `text/html` content type. This works even if the client accepts `text/*` or
`*/*`, or if the Accept header is missing, in which case `q` defaults to 1.

* It can also check the content type of the request.

Calling `n.ContentType("application/json", "text/html")` will return `true` if
the Content-Type header was set to either `application/json` or `text/html`. It
will also work if the header includes the charset, e.g. with `application/json;
charset=UTF-8`.


#### The `strict.ContentType` handler factory

In the above example, we add `strict.ContentType("application/json",
"text/xml", "")` to the list of handlers when calling `m.Post(…)`.
This will create a handler that will write a *415 Unsupported Media Type*
response if the request content type is none of the acceptable content types
passed in as arguments. The empty string means that we also want to accept
requests with no `Content-Type` header set.

`ContentType` only acts on `POST`, `PATCH` and `PUT` requests, so it is safe to
use it with `m.Use`. It will never block `GET` requests.


#### The `strict.ContentCharset` handler factory

This handler allows you to white-list a specific set of charsets for the
Content-Type header (i.e. the "UTF-8" in "Content-Type: text/plain;
charset=UTF-8"). A mismatch will result in a *415 Unsupported Media Type*
response. Including the empty string will allow headers with no charset
specified (since in most cases it is safe to assume UTF-8).

Like `ContentType`, `ContentCharset` also only acts on `POST`, `PATCH` and
`PUT` requests.


#### The `strict.Accept` handler factory

In the above example, we add `strict.Accept("application/json", "text/html")`
to the list of handlers when calling `m.Get(…)`. This will create a handler
that will write a *406 Not Acceptable* response if the Accept request header
does not permit any of the supported content types we pass in as arguments.


#### The `strict.MethodNotAllowed` handler

Passing `strict.MethodNotAllowed` to `m.Router.NotFound(…)` will tell Martini
to return a *405 Method Not Allowed* instead of a *404 Not Found* when an route
is called with a method that it is not registered with. In addition to the
response status code, the Allow header will be set to the list of allowed
methods.

Note that `strict.MethodNotAllowed` will never write a 404 response, so you
have to also add a not found handler when calling `m.Router.NotFound(…)`.
A good candidate for that is `strict.NotFound`.


#### The `strict.NotFound` handler

It is similar to `http.NotFound`, except it does not write a response body. The
response will be an empty *404 Not Found* response.


#### Constants


The following constant is included for convenience:

```go
const StatusUnprocessableEntity = 422
```
