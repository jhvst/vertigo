// This file contains bunch of miscful helper functions.
// The functions here are either too rare to be assiociated to some known file
// or are met more or less everywhere across the code.
package misc

import (
	"bufio"
	"bytes"
	"net/http"
	"strings"
	"time"

	. "github.com/9uuso/vertigo/settings"

	"github.com/go-martini/martini"
	"github.com/kennygrant/sanitize"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
)

// NotFound is a shorthand JSON response for HTTP 404 errors.
func NotFound() map[string]interface{} {
	return map[string]interface{}{"error": "Not found"}
}

// TimeOffset returns timezone offset of loc in seconds from UTC.
// Loc should be valid IANA timezone location.
func TimeOffset(loc string) (int, error) {
	var timeOffset int
	l, err := time.LoadLocation(loc)
	if err != nil {
		return timeOffset, err
	}
	now := time.Now().In(l)
	_, timeOffset = now.Zone()
	return timeOffset, nil
}

// Excerpt generates 15 word excerpt from input.
// Used to make shorter summaries from blog posts.
func Excerpt(input string) string {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)
	count := 0
	var excerpt bytes.Buffer
	for scanner.Scan() && count < 15 {
		count++
		excerpt.WriteString(scanner.Text() + " ")
	}
	return sanitize.HTML(strings.TrimSpace(excerpt.String()))
}

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
func SessionRedirect(res http.ResponseWriter, req *http.Request, session sessions.Session) {
	if sessionIsAlive(session) {
		http.Redirect(res, req, "/user", 302)
	}
}

// ProtectedPage makes sure that the user is logged in. Use on pages which need authentication
// or which have to deal with user structure later on.
func ProtectedPage(req *http.Request, session sessions.Session, render render.Render) {
	if !sessionIsAlive(session) {
		session.Delete("user")
		render.JSON(401, map[string]interface{}{"error": "Unauthorized"})
	}
}

// root returns HTTP request "root".
// For example, calling it with http.Request which has URL of /api/user/5348482a2142dfb84ca41085
// would return "api". This function is used to route both JSON API and frontend requests in the same function.
func Root(r *http.Request) string {
	u := strings.TrimPrefix(r.URL.String(), Settings.URL.Path)
	return strings.Split(u[1:], "/")[0]
}
