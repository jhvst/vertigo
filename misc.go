package main

// This file contains bunch of miscful helper functions.
// The functions here are either too rare to be assiociated to some known file
// or are met more or less everywhere across the code.
// So they are basically orphans and heroes at the same time.

import (
	r "github.com/dancannon/gorethink"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
	"net/http"
	"os"
	"strings"
	"time"
)

// Middleware function hooks the RethinkDB to be accessible for Martini routes.
// By default the middleware spawns a session pool of 10 connections.
// Typical connection options on development environment would be
//		Address: "localhost:28015"
//		Database: "test"
func middleware() martini.Handler {
	session, err := r.Connect(r.ConnectOpts{
		Address:     os.Getenv("RDB_HOST")+":"+os.Getenv("RDB_PORT"),
		Database:    "vertigo",
		MaxIdle:     10,
		IdleTimeout: time.Second * 10,
	})

	if err != nil {
		panic(err)
	}

	return func(c martini.Context) {
		c.Map(session)
	}
}

// sessionIsAlive checks that session cookie with label "user" exists and is valid.
func sessionIsAlive(session sessions.Session) bool {
	data := session.Get("user")
	_, exists := data.(string)
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
func ProtectedPage(res http.ResponseWriter, req *http.Request, session sessions.Session) {
	if !sessionIsAlive(session) {
		session.Delete("user")
		http.Redirect(res, req, "/", 302)
	}
}

// root returns HTTP request "root".
// For example, calling it with http.Request which has URL of /api/user/5348482a2142dfb84ca41085
// would return "api". This function is used to route both JSON API and frontend requests in the same function.
func root(req *http.Request) string {
	return strings.Split(strings.TrimPrefix(req.URL.String(), "/"), "/")[0]
}
