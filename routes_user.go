package main

import (
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"labix.org/v2/mgo"
)

func RoutesUser(s sessions.Session, r render.Render, db *mgo.Database) {
	person, err := GetUserFromSession(db, s)
	if err != nil {
		r.HTML(500, "error", err)
		return
	}
	posts, err := GetPostsFromAuthor(db, person)
	if err != nil {
		r.HTML(500, "error", err)
		return
	}
	r.HTML(200, "user/index", person)
	r.HTML(200, "user/listPost", posts)
}

func Register(s sessions.Session, r render.Render, db *mgo.Database, person Person) {
	person.Digest = GenerateHash(person.Password)
	s.Set("user", person.Email)
	db.C("users").Insert(person)
}
