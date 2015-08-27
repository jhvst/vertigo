package main

import (
	"net/http"
	"runtime"
	"time"

	. "github.com/9uuso/vertigo/databases/gorm"
	. "github.com/9uuso/vertigo/misc"
	. "github.com/9uuso/vertigo/routes"
	. "github.com/9uuso/vertigo/settings"

	"vertigo/render"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/sessions"
	"github.com/martini-contrib/strict"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // defining gomaxprocs is proven to add performance by few percentages
}

// NewServer spaws a new Vertigo server
func NewServer() *martini.ClassicMartini {

	m := martini.Classic()
	store := sessions.NewCookieStore([]byte(Settings.CookieHash))
	m.Use(sessions.Sessions("user", store))
	m.Use(Sessionchecker())
	m.Use(strict.Strict)
	m.Use(martini.Static("public", martini.StaticOptions{
		SkipLogging: true,
		// Adds 7 day Expire header for static files.
		Expires: func() string {
			return time.Now().Add(time.Hour * 168).UTC().Format("Mon, Jan 2 2006 15:04:05 GMT")
		},
	}))

	m.Get("/", Homepage)

	m.Get("/rss", ReadFeed)

	m.Group("/post", func(r martini.Router) {

		// Please note that `/new` route has to be before the `/:slug` route. Otherwise the program will try
		// to fetch for Post named "new".
		// For now I'll keep it this way to streamline route naming.
		r.Get("/new", ProtectedPage, func(w http.ResponseWriter) {
			render.R.HTML(w, 200, "post/new", nil)
		})
		r.Get("/:slug", ReadPost)
		r.Get("/:slug/edit", ProtectedPage, EditPost)
		r.Post("/:slug/edit", ProtectedPage, strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Post{}), binding.ErrorHandler, UpdatePost)
		r.Get("/:slug/delete", ProtectedPage, DeletePost)
		r.Get("/:slug/publish", ProtectedPage, PublishPost)
		r.Get("/:slug/unpublish", ProtectedPage, UnpublishPost)
		r.Post("/new", ProtectedPage, strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Post{}), binding.ErrorHandler, CreatePost)
		r.Post("/search", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Search{}), binding.ErrorHandler, SearchPost)

	})

	m.Group("/user", func(r martini.Router) {

		r.Get("", ProtectedPage, ReadUser)
		//r.Post("/delete", strict.ContentType("application/x-www-form-urlencoded"), ProtectedPage, binding.Form(User{}), DeleteUser)

		r.Get("/settings", ProtectedPage, ReadSettings)
		r.Post("/settings", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Vertigo{}), binding.ErrorHandler, ProtectedPage, UpdateSettings)

		r.Post("/installation", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(Vertigo{}), binding.ErrorHandler, UpdateSettings)

		r.Get("/register", SessionRedirect, func(w http.ResponseWriter) {
			render.R.HTML(w, 200, "user/register", nil)
		})
		r.Post("/register", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(User{}), binding.ErrorHandler, CreateUser)

		r.Get("/recover", SessionRedirect, func(w http.ResponseWriter) {
			render.R.HTML(w, 200, "user/recover", nil)
		})
		r.Post("/recover", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(User{}), RecoverUser)
		r.Get("/reset/:id/:recovery", SessionRedirect, func(w http.ResponseWriter) {
			render.R.HTML(w, 200, "user/reset", nil)
		})
		r.Post("/reset/:id/:recovery", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(User{}), ResetUserPassword)

		r.Get("/login", SessionRedirect, func(w http.ResponseWriter) {
			render.R.HTML(w, 200, "user/login", nil)
		})
		r.Post("/login", strict.ContentType("application/x-www-form-urlencoded"), binding.Form(User{}), LoginUser)
		r.Get("/logout", LogoutUser)

	})

	m.Group("/api", func(r martini.Router) {

		r.Get("", func(w http.ResponseWriter) {
			render.R.HTML(w, 200, "api/index", nil)
		})
		r.Get("/settings", ProtectedPage, ReadSettings)
		r.Post("/settings", strict.ContentType("application/json"), binding.Json(Vertigo{}), binding.ErrorHandler, ProtectedPage, UpdateSettings)
		r.Post("/installation", strict.ContentType("application/json"), binding.Json(Vertigo{}), binding.ErrorHandler, UpdateSettings)
		r.Get("/users", ReadUsers)
		r.Get("/user/logout", LogoutUser)
		r.Get("/user/:id", ReadUser)
		//r.Delete("/user", DeleteUser)
		r.Post("/user", strict.ContentType("application/json"), binding.Json(User{}), binding.ErrorHandler, CreateUser)
		r.Post("/user/login", strict.ContentType("application/json"), binding.Json(User{}), binding.ErrorHandler, LoginUser)
		r.Post("/user/recover", strict.ContentType("application/json"), binding.Json(User{}), binding.ErrorHandler, RecoverUser)
		r.Post("/user/reset/:id/:recovery", strict.ContentType("application/json"), binding.Json(User{}), ResetUserPassword)

		r.Get("/posts", ReadPosts)
		r.Get("/post/:slug", ReadPost)
		r.Post("/post", strict.ContentType("application/json"), binding.Json(Post{}), binding.ErrorHandler, ProtectedPage, CreatePost)
		r.Get("/post/:slug/publish", ProtectedPage, PublishPost)
		r.Get("/post/:slug/unpublish", ProtectedPage, UnpublishPost)
		r.Post("/post/:slug/edit", strict.ContentType("application/json"), binding.Json(Post{}), binding.ErrorHandler, ProtectedPage, UpdatePost)
		r.Get("/post/:slug/delete", ProtectedPage, DeletePost)
		r.Post("/post", strict.ContentType("application/json"), binding.Json(Post{}), binding.ErrorHandler, ProtectedPage, CreatePost)
		r.Post("/post/search", strict.ContentType("application/json"), binding.Json(Search{}), binding.ErrorHandler, SearchPost)

	})

	m.Router.NotFound(strict.MethodNotAllowed, func(w http.ResponseWriter) {
		render.R.HTML(w, 404, "404", nil)
	})

	return m
}

func main() {
	server := NewServer()
	server.Run()
}
