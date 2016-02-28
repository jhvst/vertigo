// Copyright 2015 Husobee Associates, LLC.  All rights reserved.
// Use of this source code is governed by The MIT License, which
// can be found in the LICENSE file included.

package vestigo

import "net/http"

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

// registerVar - Put the URL Parameter into the request's RawQuery, very PAT like.
func registerVar(r *http.Request, fmtpname string, pvalue string) {
	if r.URL.RawQuery != "" {
		r.URL.RawQuery += "&" + fmtpname + pvalue
	} else {
		r.URL.RawQuery += fmtpname + pvalue
	}
}
