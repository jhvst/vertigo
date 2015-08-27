// Package misc contains bunch of miscful helpers,
// which cannot easily be associated to any other package.
package misc

import (
	"net/http"
	"strings"

	. "github.com/9uuso/vertigo/settings"
	"vertigo/render"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
)

func Sessionchecker() martini.Handler {
	return func(session sessions.Session) {
		data := session.Get("user")
		_, exists := data.(int64)
		if exists {
			return
		}
		session.Set("user", -1)
		return
	}
}

// sessionIsAlive checks that session cookie with label "user" exists and is valid.
func sessionIsAlive(session sessions.Session) bool {
	data := session.Get("user")
	_, exists := data.(int64)
	if exists {
		return true
	}
	return false
}

// SessionRedirect in addition to sessionIsAlive makes HTTP redirection to user home.
// SessionRedirect is useful for redirecting from pages which are only visible when logged out,
// for example login and register pages.
func SessionRedirect(w http.ResponseWriter, r *http.Request, session sessions.Session) {
	if sessionIsAlive(session) {
		http.Redirect(w, r, "/user", 302)
	}
}

// ProtectedPage makes sure that the user is logged in. Use on pages which need authentication
// or which have to deal with user structure later on.
func ProtectedPage(w http.ResponseWriter, r *http.Request, session sessions.Session) {
	if !sessionIsAlive(session) {
		session.Delete("user")
		render.R.JSON(w, 401, map[string]interface{}{"error": "Unauthorized"})
	}
}

// root returns HTTP request "root".
// For example, calling it with http.Request which has URL of /api/user/5348482a2142dfb84ca41085
// would return "api". This function is used to route both JSON API and frontend requests in the same function.
func Root(r *http.Request) string {
	u := strings.TrimPrefix(r.URL.String(), Settings.URL.Path)
	return strings.Split(u[1:], "/")[0]
}
