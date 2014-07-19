// Package strict provides helpers for implementing strict APIs in Martini.
package strict

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-martini/martini"
)

// HTTP 422 Unprocessable Entity
const StatusUnprocessableEntity = 422

// Negotiator is an interface that can be used to negotiate the content type.
type Negotiator interface {
	Accepts(string) float64
	ContentType(...string) bool
}

// Negotiator implementation.
type negotiator struct {
	r *http.Request
}

// Accept parses the Accept header according to RFC2616 and returns the q value.
func (n *negotiator) Accepts(ctype string) float64 {
	return accepts(n.r.Header.Get("Accept"), ctype)
}

// ContentType checks if the Content-Type header matches the passed argument.
func (n *negotiator) ContentType(ctypes ...string) bool {
	return checkCT(n.r.Header.Get("Content-Type"), ctypes...)
}

// Strict is a `martini.Handler` that provides a `Negotiator` instance.
func Strict(r *http.Request, c martini.Context) {
	c.MapTo(&negotiator{r}, (*Negotiator)(nil))
}

// ContentType generates a handler that writes a 415 Unsupported Media Type response if none of the types match.
// An empty type will allow requests with empty or missing Content-Type header.
func ContentType(ctypes ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if contentMethod(r.Method) && !checkCT(r.Header.Get("Content-Type"), ctypes...) {
			w.WriteHeader(http.StatusUnsupportedMediaType)
		}
	}
}

// ContentCharset generates a handler that writes a 415 Unsupported Media Type response if none of the charsets match.
// An empty charset will allow requests with no Content-Type header or no specified charset.
func ContentCharset(charsets ...string) http.HandlerFunc {
	for i, c := range charsets {
		charsets[i] = strings.ToLower(c)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if contentMethod(r.Method) && !checkCC(r.Header.Get("Content-Type"), charsets...) {
			w.WriteHeader(http.StatusUnsupportedMediaType)
		}
	}
}

// Accept generates a handler that writes a 406 Not Acceptable response if none of the types match.
func Accept(ctypes ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := r.Header.Get("Accept")
		for _, t := range ctypes {
			if accepts(a, t) > 0 {
				return
			}
		}
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

// MethodNotAllowed writes a 405 Method Not Allowed response when applicable.
// It also sets the Accept header to the list of methods that are acceptable.
func MethodNotAllowed(routes martini.Routes, w http.ResponseWriter, r *http.Request) {
	if methods := routes.MethodsFor(r.URL.Path); len(methods) != 0 {
		w.Header().Set("Allow", strings.Join(methods, ","))
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// NotFound writes a 404 Not Found response.
// The difference between this and `http.NotFound` is that this method does not write a body.
func NotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

// RFC2616 header parser (simple version).
func accepts(a, ctype string) (q float64) {
	if a == ctype || a == "*/*" || a == "" {
		// bail out in some common cases
		return 1
	}
	cGroup, cType := split(ctype, "/")
	for _, field := range strings.Split(a, ",") {
		found, match := false, false
		for i, token := range strings.Split(field, ";") {
			if i == 0 {
				// token is "type/subtype", "type/*" or "*/*"
				aGroup, aType := split(token, "/")
				if cType == aType || aType == "*" {
					if (aGroup == "*" && aType == "*") || cGroup == aGroup {
						// token matches, continue to look for a q value
						found = true
						continue
					}
				}
				break
			}
			// token is "key=value"
			k, v := split(token, "=")
			if k != "q" {
				continue
			}
			// k is "q"
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				break
			}
			if q < f {
				q = f
			}
			match = true
			break
		}
		if found && !match {
			q = 1
			break
		}
	}
	return
}

// Check the content type against a list of acceptable values.
func checkCT(ct string, ctypes ...string) bool {
	ct, _ = split(ct, ";")
	for _, t := range ctypes {
		if ct == t {
			return true
		}
	}
	return false
}

// Check the content encoding against a list of acceptable values.
func checkCC(ce string, charsets ...string) bool {
	_, ce = split(strings.ToLower(ce), ";")
	_, ce = split(ce, "charset=")
	ce, _ = split(ce, ";")
	for _, c := range charsets {
		if ce == c {
			return true
		}
	}
	return false
}

// Check if the request method can contain a content
func contentMethod(m string) bool {
	// No Content-Type for GET, HEAD, OPTIONS, DELETE and CONNECT requests.
	return m == "POST" || m == "PATCH" || m == "PUT"
}

// Split a string in two parts, cleaning any whitespace.
func split(str, sep string) (a, b string) {
	parts := strings.SplitN(str, sep, 2)
	a = strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		b = strings.TrimSpace(parts[1])
	}
	return
}
