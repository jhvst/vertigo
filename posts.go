package main

// This file contains about everything related to posts. At the top you will find routes
// and at the bottom you can find CRUD options. The functions in this file are analogous
// to the ones in users.go, although some differences exist.

import (
	"bufio"
	"bytes"
	"errors"
	r "github.com/dancannon/gorethink"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"log"
	"net/http"
	"strings"
	"time"
)

type Post struct {
	Date    int32  `json:"date" gorethink:"date"`
	Title   string `json:"title"form:"title" binding:"required" gorethink:"title"`
	Author  string `json:"author,omitempty" gorethink:"author"`
	Content string `json:",omitempty" form:"content" binding:"required" gorethink:"content"`
	Excerpt string `json:"excerpt" gorethink:"excerpt"`
}

// Generates 15 word excerpt from given input.
func Excerpt(input string) string {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)
	count := 0
	var excerpt bytes.Buffer
	for scanner.Scan() && count < 15 {
		count++
		excerpt.WriteString(scanner.Text() + " ")
	}
	return strings.TrimSuffix(excerpt.String(), " ")
}

func CreatePost(req *http.Request, s sessions.Session, db *r.Session, res render.Render, post Post) {
	entry, err := post.Insert(db, s)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch root(req) {
	case "api":
		res.JSON(200, entry)
		return
	case "post":
		res.Redirect("/post/"+entry.Title, 302)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

func ReadPosts(res render.Render, db *r.Session) {
	var post Post
	posts, err := post.GetAll(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	res.JSON(200, posts)
}

func ReadPost(req *http.Request, params martini.Params, res render.Render, db *r.Session) {
	var post Post
	if params["title"] == "new" {
		res.JSON(406, map[string]interface{}{"error": "You cant name a post with colliding route name!"})
		return
	}
	post.Title = params["title"]
	post, err := post.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	switch root(req) {
	case "api":
		res.JSON(200, post)
		return
	case "post":
		res.HTML(200, "post/display", post)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// Inserts Post object to database. The function fetches user data from session and
// after that it fills Author, Date and Excerpt fields.
func (post Post) Insert(db *r.Session, s sessions.Session) (Post, error) {
	var person Person
	person, err := person.Session(db, s)
	if err != nil {
		return post, err
	}
	post.Author = person.Id
	post.Date = int32(time.Now().Unix())
	post.Excerpt = Excerpt(post.Content)
	row, err := r.Table("posts").Insert(post).RunRow(db)
	if err != nil {
		return post, err
	}
	err = row.Scan(&post)
	if err != nil {
		return post, err
	}
	return post, err
}

func (post Post) Get(s *r.Session) (Post, error) {
	row, err := r.Table("posts").Filter(func(this r.RqlTerm) r.RqlTerm {
		return this.Field("title").Eq(post.Title)
	}).RunRow(s)
	if err != nil {
		return post, err
	}
	if row.IsNil() {
		return post, errors.New("Nothing was found.")
	}
	err = row.Scan(&post)
	if err != nil {
		return post, err
	}
	return post, err
}

func (post Post) GetAll(s *r.Session) ([]Post, error) {
	var posts []Post
	rows, err := r.Table("posts").Run(s)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err := rows.Scan(&post)
		post, err := post.Get(s)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
