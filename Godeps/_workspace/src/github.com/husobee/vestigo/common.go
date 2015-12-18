// Copyright 2015 Husobee Associates, LLC.  All rights reserved.
// Use of this source code is governed by The MIT License, which
// can be found in the LICENSE file included.

package vestigo

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// methods - a list of methods that are allowed
var methods = []string{
	"CONNECT",
	"DELETE",
	"GET",
	"HEAD",
	"OPTIONS",
	"PATCH",
	"POST",
	"PUT",
	"TRACE",
}

// AllowTrace - Globally allow the TRACE method handling within vestigo url router.  This
// generally not a good idea to have true in production settings, but excellent for testing.
var AllowTrace = false

// Get - Helper method to add HTTP GET Method to router
func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.Add("GET", path, handler)
}

// Post - Helper method to add HTTP POST Method to router
func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.Add("POST", path, handler)
}

// Connect - Helper method to add HTTP CONNECT Method to router
func (r *Router) Connect(path string, handler http.HandlerFunc) {
	r.Add("CONNECT", path, handler)
}

// Delete - Helper method to add HTTP DELETE Method to router
func (r *Router) Delete(path string, handler http.HandlerFunc) {
	r.Add("DELETE", path, handler)
}

// Patch - Helper method to add HTTP PATCH Method to router
func (r *Router) Patch(path string, handler http.HandlerFunc) {
	r.Add("PATCH", path, handler)
}

// Put - Helper method to add HTTP PUT Method to router
func (r *Router) Put(path string, handler http.HandlerFunc) {
	r.Add("PUT", path, handler)
}

// Trace - Helper method to add HTTP TRACE Method to router
func (r *Router) Trace(path string, handler http.HandlerFunc) {
	r.Add("TRACE", path, handler)
}

// Param - Get a url parameter by name
func Param(r *http.Request, name string) string {
	return r.FormValue(":" + name)
}

// ParamNames - Get a url parameter name list
func ParamNames(r *http.Request) []string {
	r.ParseForm()
	names := []string{}
	for k := range r.Form {
		names = append(names, k)
	}
	return names
}

//validMethod - validate that the http method is valid.
func validMethod(method string) bool {
	var ok = false
	for _, v := range methods {
		if v == method {
			ok = true
			break
		}
	}
	return ok
}

// CorsAccessControl - Default implementation of Cors
type CorsAccessControl struct {
	AllowOrigin      []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           time.Duration
	AllowMethods     []string
	AllowHeaders     []string
}

// GetAllowOrigin - returns the allow-origin string representation
func (c *CorsAccessControl) GetAllowOrigin() []string {
	return c.AllowOrigin
}

// GetAllowCredentials - returns the allow-credentials string representation
func (c *CorsAccessControl) GetAllowCredentials() bool {
	return c.AllowCredentials
}

// GetExposeHeaders - returns the expose-headers string representation
func (c *CorsAccessControl) GetExposeHeaders() []string {
	return c.ExposeHeaders
}

// GetMaxAge - returns the max-age string representation
func (c *CorsAccessControl) GetMaxAge() time.Duration {
	return c.MaxAge
}

// GetAllowMethods - returns the allow-methods string representation
func (c *CorsAccessControl) GetAllowMethods() []string {
	return c.AllowMethods
}

// GetAllowHeaders - returns the allow-headers string representation
func (c *CorsAccessControl) GetAllowHeaders() []string {
	return c.AllowHeaders
}

// Merge - Merge the values of one CORS policy into 'this' one
func (c *CorsAccessControl) Merge(c2 *CorsAccessControl) *CorsAccessControl {
	result := new(CorsAccessControl)
	if c != nil {
		if c2 == nil {
			result.AllowOrigin = c.GetAllowOrigin()
			result.AllowCredentials = c.GetAllowCredentials()
			result.ExposeHeaders = c.GetExposeHeaders()
			result.MaxAge = c.GetMaxAge()
			result.AllowMethods = c.GetAllowMethods()
			result.AllowHeaders = c.GetAllowHeaders()
			return result
		}

		if allowOrigin := c2.GetAllowOrigin(); len(allowOrigin) != 0 {
			result.AllowOrigin = append(c.GetAllowOrigin(), c2.GetAllowOrigin()...)
		} else {
			result.AllowOrigin = c.GetAllowOrigin()
		}
		if allowCredentials := c2.GetAllowCredentials(); allowCredentials == true {
			result.AllowCredentials = c2.GetAllowCredentials()
		} else {
			result.AllowCredentials = c.GetAllowCredentials()
		}
		if exposeHeaders := c2.GetExposeHeaders(); len(exposeHeaders) != 0 {
			h := append(c.GetExposeHeaders(), c2.GetExposeHeaders()...)
			seen := map[string]bool{}
			for i, x := range h {
				if seen[strings.ToLower(x)] {
					continue
				}
				seen[strings.ToLower(x)] = true
				result.ExposeHeaders = append(result.ExposeHeaders, h[i])
			}
		} else {
			result.ExposeHeaders = c.GetExposeHeaders()
		}
		if maxAge := c2.GetMaxAge(); maxAge.Seconds() != 0 {
			result.MaxAge = c2.GetMaxAge()
		} else {
			result.MaxAge = c.GetMaxAge()
		}
		if allowMethods := c2.GetAllowMethods(); len(allowMethods) != 0 {
			h := append(c.GetAllowMethods(), allowMethods...)
			seen := map[string]bool{}
			for i, x := range h {
				if seen[x] {
					continue
				}
				seen[x] = true
				result.AllowMethods = append(result.AllowMethods, h[i])
			}
		} else {
			result.AllowMethods = c.GetAllowMethods()
		}
		if allowHeaders := c2.GetAllowHeaders(); len(allowHeaders) != 0 {
			h := append(c.GetAllowHeaders(), c2.GetAllowHeaders()...)
			seen := map[string]bool{}
			for i, x := range h {
				if seen[strings.ToLower(x)] {
					continue
				}
				seen[strings.ToLower(x)] = true
				result.AllowHeaders = append(result.AllowHeaders, h[i])
			}
		} else {
			result.AllowHeaders = c.GetAllowHeaders()
		}
	}
	return result
}

// corsPreflight - perform CORS preflight against the CORS policy for a given resource
func corsPreflight(gcors *CorsAccessControl, lcors *CorsAccessControl, allowedMethods string, w http.ResponseWriter, r *http.Request) error {

	cors := gcors.Merge(lcors)

	if origin := r.Header.Get("Origin"); cors != nil && origin != "" {
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

		if !allowedOrigin {
			// other option headers needed
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(""))
			return errors.New("quick cors end")

		}

		// if the request includes access-control-request-method
		if method := r.Header.Get("Access-Control-Request-Method"); method != "" {
			// if there are no cors settings for this resource, use the allowedMethods,
			// if there are settings for cors, use those
			responseMethods := []string{}
			if methods := cors.GetAllowMethods(); len(methods) != 0 {
				for _, x := range methods {
					if x == method {
						responseMethods = append(responseMethods, x)
					}
				}
			} else {
				for _, x := range strings.Split(allowedMethods, ", ") {
					if x == method {
						responseMethods = append(responseMethods, x)
					}
				}
			}
			if len(responseMethods) > 0 {
				w.Header().Add("Access-Control-Allow-Methods", strings.Join(responseMethods, ", "))
			} else {
				// other option headers needed
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(""))
				return errors.New("quick cors end")
			}
		}

		// if allow credentials is allowed on this resource respond with true
		if allowCredentials := cors.GetAllowCredentials(); allowedOriginExact && allowCredentials {
			w.Header().Add("Access-Control-Allow-Credentials", "true")
		}

		if exposeHeaders := cors.GetExposeHeaders(); len(exposeHeaders) != 0 {
			// if we have expose headers, send them
			w.Header().Add("Access-Control-Expose-Headers", strings.Join(exposeHeaders, ", "))
		}
		if maxAge := cors.GetMaxAge(); maxAge.Seconds() != 0 {
			// optional, if we have a max age, send it
			sec := fmt.Sprint(int64(maxAge.Seconds()))
			w.Header().Add("Access-Control-Max-Age", sec)
		}

		if header := r.Header["Access-Control-Request-Headers"]; len(header) != 0 {
			allowHeaders := cors.GetAllowHeaders()

			goodHeaders := []string{}

			for _, x := range header {
				for _, y := range allowHeaders {
					if strings.ToLower(x) == strings.ToLower(y) {
						goodHeaders = append(goodHeaders, x)
					}
				}
			}

			if len(goodHeaders) == len(header) {
				w.Header().Add("Access-Control-Allow-Headers", strings.Join(goodHeaders, ", "))
			}
		}
	}
	return nil
}

var (
	// traceHandler - Generic Trace Handler to echo back input
	traceHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "message/http")
		w.WriteHeader(http.StatusOK)
		defer r.Body.Close()
		io.Copy(w, r.Body)
	}
	// headHandler - Generic Trace Handler to echo back input
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
)

// registerVar - Put the URL Parameter into the request's RawQuery, very PAT like.
func registerVar(r *http.Request, fmtpname string, pvalue string) {
	if r.URL.RawQuery != "" {
		r.URL.RawQuery += "&" + fmtpname + pvalue
	} else {
		r.URL.RawQuery += fmtpname + pvalue
	}
}
