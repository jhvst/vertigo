// This file contains bunch of miscful helper functions.
// The functions here are either too rare to be assiociated to some known file
// or are met more or less everywhere across the code.
package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	_ "github.com/mattn/go-sqlite3"
)

func NotFound() map[string]interface{} {
	return map[string]interface{}{"error": "Not found"}
}

func init() {

	if os.Getenv("DATABASE_URL") != "" {
		os.Setenv("driver", "postgres")
		os.Setenv("dbsource", os.Getenv("DATABASE_URL"))
		log.Println("Using PostgreSQL")
	} else {
		os.Setenv("driver", "sqlite3")
		os.Setenv("dbsource", "./vertigo.db")
		log.Println("Using SQLite3")
	}

	db, err := gorm.Open(os.Getenv("driver"), os.Getenv("dbsource"))

	if err != nil {
		panic(err)
	}

	db.LogMode(false)

	// Here database and tables are created in case they do not exist yet.
	// If database or tables do exist, nothing will happen to the original ones.
	db.CreateTable(&User{})
	db.CreateTable(&Post{})
	db.AutoMigrate(&User{}, &Post{})
}

func sessionchecker() martini.Handler {
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

// Middleware function hooks the database to be accessible for Martini routes.
func middleware() martini.Handler {
	db, err := gorm.Open(os.Getenv("driver"), os.Getenv("dbsource"))
	db.LogMode(false)
	if err != nil {
		panic(err)
	}
	return func(c martini.Context) {
		c.Map(&db)
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
func root(req *http.Request) string {
	return strings.Split(strings.TrimPrefix(req.URL.String(), "/"), "/")[0]
}
