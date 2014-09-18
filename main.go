// Main.go contains settings related to the web server, such as
// template helper functions, HTTP routes and Martini settings.
package main

import (
	"html"
	"html/template"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/martini-contrib/strict"
)

func main() {

	helpers := template.FuncMap{
		// Unescape unescapes and parses HTML from database objects.
		// Used in templates such as "/post/display.tmpl"
		"unescape": func(s string) template.HTML {
			return template.HTML(html.UnescapeString(s))
		},
		// Title renders post name as a page title.
		// Otherwise it defaults to Vertigo.
		"title": func(t interface{}) string {
			post, exists := t.(Post)
			if exists {
				return post.Title
			}
			return Settings.Name
		},
		// Date helper returns unix date as more readable one in string format.
		"date": func(d int64) string {
			return time.Unix(d, 0).String()
		},
	}

	m := martini.Classic()
	store := sessions.NewCookieStore([]byte(Settings.CookieHash))
	m.Use(sessions.Sessions("user", store))
	m.Use(middleware())
	m.Use(strict.Strict)
	m.Use(martini.Static("public", martini.StaticOptions{
		SkipLogging: true,
		Expires: func() string {
			return "Cache-Control: max-age=31536000"
		},
	}))
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
		Funcs:  []template.FuncMap{helpers}, // Specify helper function maps for templates to access.
	}))

	m.Get("/", Homepage)

	m.Group("/feeds", func(r martini.Router) {
		r.Get("", func(res render.Render) {
			res.Redirect("/feeds/rss", 302)
		})
		r.Get("/atom", ReadFeed)
		r.Get("/rss", ReadFeed)
	})

	m.Group("/post", func(r martini.Router) {

		// Please note that `/new` route has to be before the `/:title` route. Otherwise the program will try
		// to fetch for Post named "new".
		// For now I'll keep it this way to streamline route naming.
		r.Get("/new", ProtectedPage, func(res render.Render) {
			res.HTML(200, "post/new", nil)
		})
		r.Get("/:title", ReadPost)
		r.Get("/:title/edit", EditPost)
		r.Post("/:title/edit", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Post{}), binding.ErrorHandler, UpdatePost)
		r.Get("/:title/delete", DeletePost)
		r.Get("/:title/publish", PublishPost)
		r.Post("/new", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Post{}), binding.ErrorHandler, CreatePost)
		r.Post("/search", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Search{}), binding.ErrorHandler, SearchPost)

	})

	m.Group("/user", func(r martini.Router) {

		r.Get("", ProtectedPage, ReadUser)
		//r.Post("/delete", strict.ContentType("application/x-www-form-urlencoded"), ProtectedPage, binding.Form(Person{}), DeleteUser)

		m.Post("/installation", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Vertigo{}), binding.ErrorHandler, UpdateSettings)

		r.Get("/register", SessionRedirect, func(res render.Render) {
			res.HTML(200, "user/register", nil)
		})
		r.Post("/register", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Person{}), binding.ErrorHandler, CreateUser)

		r.Get("/recover", SessionRedirect, func(res render.Render) {
			res.HTML(200, "user/recover", nil)
		})
		r.Post("/recover", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Person{}), RecoverUser)
		r.Get("/reset/:id/:recovery", SessionRedirect, func(res render.Render) {
			res.HTML(200, "user/reset", nil)
		})
		r.Post("/reset/:id/:recovery", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Person{}), ResetUserPassword)

		r.Get("/login", SessionRedirect, func(res render.Render) {
			res.HTML(200, "user/login", nil)
		})
		r.Post("/login", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Person{}), LoginUser)
		r.Get("/logout", LogoutUser)

	})

	m.Group("/api", func(r martini.Router) {

		r.Get("", func(res render.Render) {
			res.HTML(200, "api/index", nil)
		})
		r.Get("/users", ReadUsers)
		r.Get("/user/:id", ReadUser)
		//r.Delete("/user", DeleteUser)
		r.Post("/user", strict.ContentType("application/json"), binding.Json(Person{}), binding.ErrorHandler, CreateUser)
		r.Post("/user/login", strict.ContentType("application/json"), binding.Json(Person{}), binding.ErrorHandler, LoginUser)
		r.Get("/user/logout", LogoutUser)

		r.Get("/posts", ReadPosts)
		r.Get("/post/:title", ReadPost)
		r.Post("/post", strict.ContentType("application/json"), binding.Json(Post{}), binding.ErrorHandler, CreatePost)
		r.Get("/post/:title/publish")
		r.Post("/post/:title/edit", strict.ContentType("application/json"), binding.Json(Post{}), binding.ErrorHandler, UpdatePost)
		r.Get("/post/:title/delete", DeletePost)
		r.Post("/post", strict.ContentType("application/json"), binding.Json(Post{}), binding.ErrorHandler, CreatePost)
		r.Post("/post/search/:query", strict.ContentType("application/json"), binding.Json(Search{}), binding.ErrorHandler, SearchPost)

	})

	m.Router.NotFound(strict.MethodNotAllowed, strict.NotFound)
	m.Run()

}
