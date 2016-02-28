# Vestigo - A Standalone Golang URL Router

[![Join the chat at https://gitter.im/husobee/vestigo](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/husobee/vestigo?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Build Status](https://travis-ci.org/husobee/vestigo.svg)](https://travis-ci.org/husobee/vestigo)
[![Coverage Status](https://coveralls.io/repos/husobee/vestigo/badge.svg?branch=master&service=github)](https://coveralls.io/github/husobee/vestigo?branch=master)
[![GoDoc](https://godoc.org/github.com/husobee/vestigo?status.svg)](https://godoc.org/github.com/husobee/vestigo)

## Abstract

Many fast Golang URL routers are often embedded inside frameworks.  Vestigo is a stand alone url router
which has respectable performance that passes URL parameters to handlers by embedding them into the request's 
Form. 

There is such an abundance of parts and pieces that can be fit together for go web services, it seems like a 
shame to have a very fast URL router require the use of one framework, and one context model.  This library 
aims to give the world a fast, and featureful URL router that can stand on it's own, without being forced into 
a particular web framework.

## Design

1. Radix Tree Based
2. Attach URL Parameters into Request (PAT style) instead of context
3. HTTP Compliance (TRACE, OPTIONS, HEAD)
4. CORS Enabled (per resource access-controls)

### TODOs for V1

- [x] Router functioning with a resource concept attached to leaf nodes
- [x] Use resources to drive responses to particular Methods (not found v not allowed)
- [x] Implement Resource and Globally scoped CORS preflights
- [x] Fix bug in router where handler.allowedMethods is getting populated where it shouldn't be
- [x] Validate with Tests RFC 2616 Compliance (OPTIONS, etc)

### TODOs for V2
- [ ] Valiators for URL params
- [ ] Implement RFC 6570 URI Parameters

## Performance

Initial implementation on a fork of [standard http performance testing libary][http-perf-test] shows the following:

```
BenchmarkVestigo_GithubAll         20000             75763 ns/op            9280 B/op        339 allocs/op
```

I should mention that the above performance is about 2x slower then the fastest URL router I have tested (Echo/Gin), and
is slightly worse than HTTPRouter, but I am happy with this performance considering this implementation is the fastest 
implementation that can handle standard http.HandlerFunc handlers, without forcing end users to use a particular context, 
or use a non-standard handler function, locking them into an implementation.

## Examples

```go

package main

import (
	"log"
	"net/http"

	"github.com/husobee/vestigo"
)

func main () {
    router := vestigo.NewRouter()
    // you can enable trace by setting this to true
    vestigo.AllowTrace = true

    // Setting up router global  CORS policy
    // These policy guidelines are overriddable at a per resource level shown below
	router.SetGlobalCors(&vestigo.CorsAccessControl{
		AllowOrigin:      []string{"*", "test.com"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-Header", "X-Y-Header"},
		MaxAge:           3600 * time.Second,
		AllowHeaders:     []string{"X-Header", "X-Y-Header"},
	})


    // setting two methods on the same resource
    router.Get("/welcome", GetWelcomeHandler)
    router.Post("/welcome", PostWelcomeHandler)

    // URL parameter "name"
    router.Post("/welcome/:name", PostWelcomeHandler)

	// Catch-All methods to allow easy migration from http.ServeMux
	router.HandleFunc("/general", GeneralHandler)

    // Below Applies Local CORS capabilities per Resource (both methods covered)
    // by default this will merge the "GlobalCors" settings with the resource
    // cors settings.  Without specifying the AllowMethods, the router will 
    // accept any Request-Methods that have valid handlers associated
	router.SetCors("/welcome", &vestigo.CorsAccessControl{
		AllowMethods: []string{"GET"}, // only allow cors for this resource on GET calls
		AllowHeaders: []string{"X-Header", "X-Z-Header"}, // Allow this one header for this resource
	})


	log.Fatal(http.ListenAndServe(":1234", router))

}

func PostWelcomeHandler(w http.ResponseWriter, r *http.Request) {
    name := vestigo.Param(r, "name") // url params live in the request
    w.WriteHeader(200)
    w.Write([]byte("welcome " + name +"!"))
}

func GetWelcomeHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    w.Write([]byte("welcome!"))
}

func GeneralHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    w.Write([]byte("Gotta catch em all!"))
}

```

## Licensing

* Portions of the URL Router were taken from [Echo][echo-main] and are covered under their [License][echo-main-license].
* Portions of the Tests were taken from [julienschmidt's go-http-routing-benchmark][http-perf-test] and are covered under
this [license][http-perf-test-license]
* The rest of the implementation is covered under *The MIT License* covered under this [License][vestigo-main-license].

# Contributing

If you wish to contribute, please fork this repository, submit an issue, or pull request with your suggestions.  
_Please use gofmt and golint before trying to contribute._


[echo-main]: https://github.com/labstack/echo
[echo-main-license]: https://github.com/labstack/echo/blob/master/LICENSE
[vestigo-main-license]: https://github.com/husobee/vestigo/blob/master/LICENSE
[http-perf-test]: https://github.com/julienschmidt/go-http-routing-benchmark
[http-perf-test-license]: https://github.com/julienschmidt/go-http-routing-benchmark/blob/master/LICENSE
