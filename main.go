package main

import (
	"github.com/attilaolah/strict"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"log"
	"html/template"
)

func main() {

	helpers := template.FuncMap{
		"unescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	m := martini.Classic()
	store := sessions.NewCookieStore([]byte("heartbleed"))
	m.Use(sessions.Sessions("user", store))
	m.Use(middleware())
	m.Use(strict.Strict)
	m.Use(gzip.All())
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
		Funcs: []template.FuncMap{helpers}, // Specify helper function maps for templates to access.
	}))

	m.Get("/", Homepage)

	m.Group("/post", func(r martini.Router) {

		// Please note that `/new` route has to be before the `/:title` route. Otherwise the program will try
		// to fetch for Post named "new".
		// For now I'll keep it this way to streamline route naming.
		r.Get("/new", ProtectedPage, func(res render.Render) {
			res.HTML(200, "post/new", nil)
		})
		r.Get("/:title", ReadPost)
		r.Post("/new", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Post{}), binding.ErrorHandler, CreatePost)

	})

	m.Group("/user", func(r martini.Router) {

		r.Get("", ProtectedPage, ReadUser)
		//r.Post("/delete", strict.ContentType("application/x-www-form-urlencoded"), ProtectedPage, binding.Form(Person{}), DeleteUser)

		r.Get("/register", SessionRedirect, func(res render.Render) {
			res.HTML(200, "user/register", nil)
		})
		r.Post("/register", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Person{}), binding.ErrorHandler, CreateUser)

		r.Get("/login", SessionRedirect, func(res render.Render) {
			res.HTML(200, "user/login", nil)
		})
		r.Post("/login", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Person{}), LoginUser)

		r.Get("/logout", func(s sessions.Session, res render.Render) {
			s.Delete("user")
			res.Redirect("/", 302)
		})

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

		r.Get("/posts", ReadPosts)
		r.Get("/post/:title", ReadPost)
		r.Post("/post", strict.ContentType("application/json"), binding.Json(Post{}), binding.ErrorHandler, CreatePost)

	})

	m.Router.NotFound(strict.MethodNotAllowed, strict.NotFound)
	m.Run()

	log.Println("Server started.")
}
