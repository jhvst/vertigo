// Copyright 2015 Husobee Associates, LLC.  All rights reserved.
// Use of this source code is governed by The MIT License, which
// can be found in the LICENSE file included.

package vestigo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

var (
	notFoundOnce sync.Once
)

// CustomNotFoundHandlerFunc - Specify a Handlerfunc to use for a custom NotFound Handler.  Can only be performed once.
func CustomNotFoundHandlerFunc(f http.HandlerFunc) {
	notFoundOnce.Do(func() {
		notFoundHandler = f
	})
}

var (
	// traceHandler - Generic Trace Handler to echo back input
	traceHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "message/http")
		w.WriteHeader(http.StatusOK)
		defer r.Body.Close()
		io.Copy(w, r.Body)
	}
	// headHandler - Generic Head Handler to return header information
	headHandler = func(f http.HandlerFunc) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			fakeWriter := httptest.NewRecorder()
			f(fakeWriter, r)
			for k, v := range fakeWriter.Header() {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			w.WriteHeader(fakeWriter.Code)
			w.Write([]byte(""))
		}
	}

	// optionsHandler - Generic Options Handler to handle when method isn't allowed for a resource
	optionsHandler = func(gcors *CorsAccessControl, lcors *CorsAccessControl, allowedMethods string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Allow", allowedMethods)

			if err := corsPreflight(gcors, lcors, allowedMethods, w, r); err != nil {
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(""))
		}
	}
	// methodNotAllowedHandler - Generic Handler to handle when method isn't allowed for a resource
	methodNotAllowedHandler = func(allowedMethods string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Allow", allowedMethods)
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method Not Allowed"))
		}
	}
	// notFoundHandler - Generic Handler to handle when resource isn't found
	notFoundHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}

	// corsFlightWrapper - Wrap the handler in cors
	corsFlightWrapper = func(gcors *CorsAccessControl, lcors *CorsAccessControl, allowedMethods string, f func(http.ResponseWriter, *http.Request)) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {

			if origin := r.Header.Get("Origin"); origin != "" {
				cors := gcors.Merge(lcors)
				if cors != nil {
					// validate origin is in list of acceptable allow-origins
					allowedOrigin := false
					allowedOriginExact := false
					for _, v := range cors.GetAllowOrigin() {
						if v == origin {
							w.Header().Add("Access-Control-Allow-Origin", origin)
							allowedOriginExact = true
							allowedOrigin = true
							break
						}
					}
					if !allowedOrigin {
						for _, v := range cors.GetAllowOrigin() {
							if v == "*" {
								w.Header().Add("Access-Control-Allow-Origin", v)
								allowedOrigin = true
								break
							}
						}
					}

					// if allow credentials is allowed on this resource respond with true
					if allowCredentials := cors.GetAllowCredentials(); allowedOriginExact && allowCredentials {
						w.Header().Add("Access-Control-Allow-Credentials", "true")
					}
				}
			}
			f(w, r)
		}
	}
)
