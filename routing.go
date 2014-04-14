package main

import (
	"github.com/go-martini/martini"
	"net/http"
	"strings"
)

//Returns which database collection to use based on request path.
//Example: RouteDomain(/api/user/5348482a2142dfb84ca41085) returns "user"
func RouteDomain(req *http.Request, params martini.Params) string {
	for _, path := range params {
		u := strings.TrimSuffix(req.URL.Path, "/"+path)
		r := strings.Split(u, "/")
		method := r[len(r)-1]
		return method
	}
	return ""
}
