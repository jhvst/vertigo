package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"time"
)

type Person struct {
	Id       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name     string        `json:"name" form:"name" binding:"required"`
	Digest   []byte
	//the password field should not be saved to database - should there be seperate structures for forms and actual db objects?
	Password string `form:"password"`
	Email    string `json:"email" form:"email" binding:"required"`
	Posts Post
}

type Post struct {
	Date int32
	Title string `form:"title" binding:"required"`
	Author string
	Content string `form:"content" binding:"required"`
}

func (ps Person) Validate(errors *binding.Errors, req *http.Request) {
	if EmailIsUnique(ps) != true {
		errors.Fields["email"] = "Email is already in use."
	}
}

func SessionIsAlive(session sessions.Session) bool {
	data := session.Get("user")
	_, exists := data.(string)
	if exists {
		return true
	}
	return false
}

func SessionRedirect(c martini.Context, res http.ResponseWriter, req *http.Request, r render.Render, session sessions.Session) {
	if SessionIsAlive(session) {
		http.Redirect(res, req, "/user", 302)
	}
}

func HomeRedirect(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/", 302)
}

func ClearSessionCookies(session sessions.Session) {
	session.Clear()
}

func ProtectedPage(res http.ResponseWriter, req *http.Request, session sessions.Session) {
	if !SessionIsAlive(session) {
		session.Clear()
		http.Redirect(res, req, "/", 302)
	}
}

func main() {
	m := martini.Classic()
	store := sessions.NewCookieStore([]byte("heartbleed"))
	m.Use(sessions.Sessions("user", store))
	m.Use(render.Renderer())
	m.Use(MongoMiddleware())
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "home", nil)
	})

	m.Get("/users", func(r render.Render, db *mgo.Database) {
		r.HTML(200, "users", GetAll(db))
	})

	m.Get("/post/new", ProtectedPage, func(r render.Render, db *mgo.Database) {
		r.HTML(200, "post/new", nil)
	})

	m.Get("/post/:title", func(params martini.Params, r render.Render, db *mgo.Database) {
		var post Post
		post.Title = params["title"]
		post, err := GetPostWithTitle(db, &post)
		if err != nil {
			r.HTML(500, "error", err)
		}
		r.HTML(200, "post/display", post)
	})

	m.Post("/post/new", ProtectedPage, binding.Form(Post{}), binding.ErrorHandler, func(session sessions.Session, r render.Render, db *mgo.Database, post Post) {
		person, err := GetUserFromSession(db, session)
		if err != nil {
			r.HTML(500, "error", err)
		}
		post.Date = int32(time.Now().Unix())
		post.Author = person.Email
		db.C("posts").Insert(post)
	}, SessionRedirect)

	m.Get("/user", ProtectedPage, func(session sessions.Session, r render.Render, db *mgo.Database) {
		person, err := GetUserFromSession(db, session)
		if err != nil {
			r.HTML(500, "error", err)
		}
		r.HTML(200, "user/index", person)
	})

	m.Get("/user/register", SessionRedirect, func(r render.Render) {
		r.HTML(200, "user/register", nil)
	})

	m.Post("/user", ProtectedPage, binding.Form(Person{}), binding.ErrorHandler, func(session sessions.Session, r render.Render, db *mgo.Database, person Person) {
		err := UpdateUserBySession(db, session, person)
		if err != nil {
			session.Clear()
			r.HTML(500, "error", err)
			return
		}
	}, SessionRedirect)

	m.Post("/user/register", binding.Form(Person{}), binding.ErrorHandler, func(s sessions.Session, r render.Render, db *mgo.Database, person Person) {
		person.Digest = GenerateHash(person.Password)
		person.Password = ""
		s.Set("user", person.Email)
		db.C("users").Insert(person)
	}, SessionRedirect)

	m.Get("/user/login", SessionRedirect, func(r render.Render) {
		r.HTML(200, "user/login", nil)
	})

	m.Get("/user/logout", ClearSessionCookies, HomeRedirect)

	m.Post("/user/login", binding.Form(Person{}), func(s sessions.Session, r render.Render, db *mgo.Database, person Person) {
		submittedPassword := person.Password
		person, err := GetUserWithEmail(db, &person)
		if err != nil {
			r.HTML(401, "user/login", "Wrong username or password.")
			return
		}
		if CompareHash(person.Digest, submittedPassword) {
			s.Set("user", person.Email)
			return
		}
		r.HTML(401, "user/login", "Wrong username or password.")
	}, SessionRedirect)

	m.Post("/user/delete", ProtectedPage, binding.Form(Person{}), func(session sessions.Session, r render.Render, db *mgo.Database, person Person) {
		submittedPassword := person.Password
		person, err := GetUserFromSession(db, session)
		if err != nil {
			r.HTML(500, "error", "Database connection error. Please try again later.")
			r.HTML(500, "user/index", person)
			return
		}
		if CompareHash(person.Digest, submittedPassword) {
			err := RemoveUserByID(db, &person)
			if err != nil {
				r.HTML(401, "error", "Wrong username or password.")
				r.HTML(401, "user/index", person)
				return
			}
			session.Clear()
		}
	//this should probably lead to somewhere where you can check whether the response actually succeeded...
	}, SessionRedirect)

	m.Get("/api/users", func(r render.Render, db *mgo.Database) {
		r.JSON(200, GetAll(db))
	})

	m.Get("/api/user/:id", func(params martini.Params, r render.Render, db *mgo.Database) {
		r.JSON(200, GetUserByID(db, params["id"]))
	})

	m.Run()

	log.Println("Server started.")
}
