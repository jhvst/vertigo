package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	. "github.com/9uuso/vertigo/databases/sqlx"
	"github.com/9uuso/vertigo/render"
	. "github.com/9uuso/vertigo/routes"
	. "github.com/9uuso/vertigo/session"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/husobee/vestigo"
	"github.com/justinas/alice"
)

var Store = sessions.NewCookieStore([]byte(Settings.CookieHash))

func session(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, "session", Store)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func bindPost(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Content-Type"][0] == "application/json" {
			var post Post
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&post)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			context.Set(r, "post", post)
			next.ServeHTTP(w, r)
			return
		}

		r.ParseForm()
		title := r.PostFormValue("title")
		if title == "" {
			http.Error(w, "Title is required.", http.StatusBadRequest)
			return
		}

		var post Post
		post.Title = title
		post.Markdown = r.PostFormValue("markdown")
		context.Set(r, "post", post)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func bindSearch(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Content-Type"][0] == "application/json" {
			var search Search
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&search)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			context.Set(r, "search", search)
			next.ServeHTTP(w, r)
			return
		}

		r.ParseForm()
		query := r.PostFormValue("query")
		if query == "" {
			http.Error(w, "Query is required.", http.StatusBadRequest)
			return
		}

		var search Search
		search.Query = query
		context.Set(r, "search", search)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func bindSettings(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		var settings Vertigo

		if r.Header["Content-Type"][0] == "application/json" {

			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&settings)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			context.Set(r, "settings", settings)
			next.ServeHTTP(w, r)
			return
		}

		r.ParseForm()

		name := r.PostFormValue("name")
		if name == "" {
			http.Error(w, "Name is required.", http.StatusBadRequest)
			return
		}

		hostname := r.PostFormValue("hostname")
		if hostname == "" {
			http.Error(w, "Hostname is required.", http.StatusBadRequest)
			return
		}

		description := r.PostFormValue("description")
		if description == "" {
			http.Error(w, "Description is required.", http.StatusBadRequest)
			return
		}

		if r.PostFormValue("allowregistrations") != "" {
			allowregistrations, err := strconv.ParseBool(r.PostFormValue("allowregistrations"))
			if err != nil {
				http.Error(w, "Allowregistrations is required.", http.StatusBadRequest)
				return
			}
			settings.AllowRegistrations = allowregistrations
		}

		if r.PostFormValue("mailerport") != "" {
			port, err := strconv.ParseInt(r.PostFormValue("mailerport"), 10, 64)
			if err != nil {
				http.Error(w, "Mailer port needs to be a number.", http.StatusBadRequest)
				return
			}
			settings.MailerPort = int(port)
		}

		settings.Name = name
		settings.Hostname = hostname
		settings.Description = description
		settings.MailerLogin = r.PostFormValue("mailerlogin")
		settings.MailerPassword = r.PostFormValue("mailerpassword")
		settings.MailerHostname = r.PostFormValue("mailerhostname")
		context.Set(r, "settings", settings)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func bindUser(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Content-Type"][0] == "application/json" {
			var user User
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			context.Set(r, "user", user)
			next.ServeHTTP(w, r)
			return
		}

		r.ParseForm()

		email := r.PostFormValue("email")
		if email == "" {
			http.Error(w, "Email is required.", http.StatusBadRequest)
			return
		}

		password := r.PostFormValue("password")
		if email == "" {
			http.Error(w, "Password is required.", http.StatusBadRequest)
			return
		}

		var user User
		user.Email = email
		user.Password = password
		user.Location = r.PostFormValue("location")
		user.Name = r.PostFormValue("name")
		user.Recovery = r.PostFormValue("recovery")
		context.Set(r, "user", user)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func bindReset(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Content-Type"][0] == "application/json" {
			var user User
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if user.Password == "" {
				http.Error(w, `{"error": "Password is required."}`, http.StatusBadRequest)
				return
			}

			context.Set(r, "newpassword", user.Password)
			next.ServeHTTP(w, r)
			return
		}

		r.ParseForm()

		password := r.PostFormValue("password")
		if password == "" {
			http.Error(w, "Password is required.", http.StatusBadRequest)
			return
		}

		context.Set(r, "newpassword", password)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func staticFile(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/"+r.URL.Path[1:])
}

func staticResource(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path
	s := http.Dir("static")

	f, err := s.Open(strings.TrimPrefix(file, "/static"))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	http.ServeContent(w, r, file, fi.ModTime(), f)
}

func NewServer() *vestigo.Router {

	protectedHandler := alice.New(session, ProtectedPage)
	postForm := alice.New(session, ProtectedPage, bindPost)
	postUser := alice.New(session, bindUser)
	recoverUser := alice.New(session, bindUser)
	postSearch := alice.New(bindSearch)
	postReset := alice.New(bindReset)
	postSettings := alice.New(session, bindSettings)
	sessionRedirect := alice.New(session, SessionRedirect)

	r := vestigo.NewRouter()

	r.Get("/", Homepage)
	r.Get("/rss", ReadFeed)
	r.Get("/apple-touch-icon.png", staticFile)
	r.Get("/favicon.ico", staticFile)
	r.Get("/browserconfig.xml", staticFile)
	r.Get("/crossdomain.xml", staticFile)
	r.Get("/robots.txt", staticFile)
	r.Get("/tile-wide.png", staticFile)
	r.Get("/tile.png", staticFile)
	r.Get("/static/*", staticResource)
	// Please note that `/new` route has to be before the `/:slug` route. Otherwise the program will try
	// to fetch for Post named "new".
	// For now I'll keep it this way to streamline route naming.
	r.Get("/posts/new", protectedHandler.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		render.R.HTML(w, 200, "post/new", nil)
	}).(http.HandlerFunc))
	r.Post("/posts/new", postForm.ThenFunc(CreatePost).(http.HandlerFunc))

	r.Post("/posts/search", postSearch.ThenFunc(SearchPost).(http.HandlerFunc))
	r.Get("/post/:slug/edit", protectedHandler.ThenFunc(EditPost).(http.HandlerFunc))
	r.Post("/post/:slug/edit", postForm.ThenFunc(UpdatePost).(http.HandlerFunc))
	r.Get("/post/:slug/delete", protectedHandler.ThenFunc(DeletePost).(http.HandlerFunc))
	r.Get("/post/:slug/publish", protectedHandler.ThenFunc(PublishPost).(http.HandlerFunc))
	r.Get("/post/:slug/unpublish", protectedHandler.ThenFunc(UnpublishPost).(http.HandlerFunc))
	r.Get("/post/:slug", ReadPost)

	r.Get("/user", protectedHandler.Then(http.HandlerFunc(ReadUser)).(http.HandlerFunc))
	//r.HandleFunc("/delete", ProtectedPage, binding.Form(User{}), DeleteUser)
	r.Get("/user/settings", protectedHandler.ThenFunc(ReadSettings).(http.HandlerFunc))
	r.Post("/user/settings", postSettings.ThenFunc(UpdateSettings).(http.HandlerFunc))

	r.Post("/user/installation", postSettings.ThenFunc(UpdateSettings).(http.HandlerFunc))

	r.Get("/user/register", sessionRedirect.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		render.R.HTML(w, 200, "user/register", nil)
	}).(http.HandlerFunc))

	r.Post("/user/register", postUser.ThenFunc(CreateUser).(http.HandlerFunc))

	r.Get("/user/recover", func(w http.ResponseWriter, r *http.Request) {
		render.R.HTML(w, 200, "user/recover", nil)
	})

	r.Post("/user/recover", postUser.ThenFunc(RecoverUser).(http.HandlerFunc))
	r.Get("/user/reset/:id/:recovery", sessionRedirect.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		render.R.HTML(w, 200, "user/reset", nil)
	}).(http.HandlerFunc))

	r.Post("/user/reset/:id/:recovery", postReset.ThenFunc(ResetUserPassword).(http.HandlerFunc))

	r.Get("/user/login", sessionRedirect.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		render.R.HTML(w, 200, "user/login", nil)
	}).(http.HandlerFunc))

	r.Post("/user/login", recoverUser.ThenFunc(LoginUser).(http.HandlerFunc))
	r.Get("/user/logout", LogoutUser)

	r.Get("/api", func(w http.ResponseWriter, r *http.Request) {
		render.R.HTML(w, 200, "api/index", nil)
	})

	r.Get("/api/settings", protectedHandler.ThenFunc(ReadSettings).(http.HandlerFunc))
	r.Post("/api/settings", postSettings.ThenFunc(UpdateSettings).(http.HandlerFunc))
	r.Post("/api/installation", postSettings.ThenFunc(UpdateSettings).(http.HandlerFunc))
	r.Get("/api/users", ReadUsers)
	r.Get("/api/users/", ReadUsers)
	r.Get("/api/user/logout", LogoutUser)
	r.Get("/api/user/:id", ReadUser)
	//r.Delete("/user", DeleteUser)
	r.Post("/api/user", postUser.ThenFunc(CreateUser).(http.HandlerFunc))
	r.Post("/api/user/login", recoverUser.ThenFunc(LoginUser).(http.HandlerFunc))
	r.Post("/api/user/recover", recoverUser.ThenFunc(RecoverUser).(http.HandlerFunc))
	r.Post("/api/user/reset/:id/:recovery", postReset.ThenFunc(ResetUserPassword).(http.HandlerFunc))

	r.Post("/api/posts/search", postSearch.ThenFunc(SearchPost).(http.HandlerFunc))
	r.Get("/api/posts", ReadPosts)
	r.Post("/api/post", postForm.ThenFunc(CreatePost).(http.HandlerFunc))
	r.Post("/api/post/:slug/edit", postForm.ThenFunc(UpdatePost).(http.HandlerFunc))
	r.Get("/api/post/:slug/delete", protectedHandler.ThenFunc(DeletePost).(http.HandlerFunc))
	r.Get("/api/post/:slug/publish", protectedHandler.ThenFunc(PublishPost).(http.HandlerFunc))
	r.Get("/api/post/:slug/unpublish", protectedHandler.ThenFunc(UnpublishPost).(http.HandlerFunc))
	r.Get("/api/post/:slug", ReadPost)

	return r
}

func main() {
	server := NewServer()
	if os.Getenv("PORT") == "" {
		log.Fatal(http.ListenAndServe(":3000", server))
	} else {
		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			log.Println("Port was defined but could not be parsed.")
			os.Exit(1)
		}
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), server))
	}
}
