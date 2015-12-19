// Package session contains session functions.
package session

import (
	"net/http"
	"net/url"
	"strings"

	. "github.com/9uuso/vertigo/databases/sqlx"
	"github.com/9uuso/vertigo/render"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
)

const key string = "vertigo"

func GetSession(r *http.Request) *sessions.CookieStore {
	if rv := context.Get(r, "session"); rv != nil {
		return rv.(*sessions.CookieStore)
	}
	return nil
}

// sessionIsAlive checks that session cookie with label "id" exists and is valid.
func sessionIsAlive(r *http.Request) bool {
	s, ok := SessionGetValue(r, "id")
	if s < 1 || ok == false {
		return false
	}
	return true
}

func SessionGetValue(r *http.Request, key string) (value int64, ok bool) {
	store := GetSession(r)
	session, _ := store.Get(r, key)
	token, ok := session.Values[key]
	if ok {
		return token.(int64), true
	}
	return 0, false
}

func SessionSetValue(w http.ResponseWriter, r *http.Request, key string, value int64) {
	store := GetSession(r)
	session, _ := store.Get(r, key)
	session.Values[key] = value
	session.Save(r, w)
}

func SessionDelete(w http.ResponseWriter, r *http.Request, key string) {
	store := GetSession(r)
	// panics unless session exists
	if store != nil {
		session, _ := store.Get(r, key)
		session.Options = &sessions.Options{Path: "/", MaxAge: -1, Secure: true, HttpOnly: true}
		session.Save(r, w)
	}
}

// SessionRedirect in addition to sessionIsAlive makes HTTP redirection to user home.
// SessionRedirect is useful for redirecting from pages which are only visible when logged out,
// for example login and register pages.
func SessionRedirect(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if sessionIsAlive(r) {
			http.Redirect(w, r, "/user", 302)
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// ProtectedPage makes sure that the user is logged in. Use on pages which need authentication
// or which have to deal with user structure later on.
func ProtectedPage(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !sessionIsAlive(r) {
			SessionDelete(w, r, "id")
			render.R.JSON(w, 401, map[string]interface{}{"error": "Unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// root returns HTTP request "root".
// For example, calling it with http.Request which has URL of /api/user/5348482a2142dfb84ca41085
// would return "api". This function is used to route both JSON API and frontend requests in the same function.
func Root(r *http.Request) string {
	su, _ := url.Parse(Settings.Hostname)
	u := strings.TrimPrefix(r.URL.String(), su.Path)
	return strings.Split(u[1:], "/")[0]
}
