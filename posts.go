package main

import (
	r "github.com/dancannon/gorethink"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	//"time"
	"errors"
	//"log"
	"bufio"
	"os"
	"strings"
)

type Post struct {
	Date    int32  `json:"date"`
	Title   string `form:"title" json:"title" binding:"required"`
	Author  string `json:"author,omitempty"`
	Content string `form:"content" binding:"required" json:"-"`
	Excerpt string `json:"excerpt"`
}

func Excerpt(input string) string {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)
	count := 0
	excerpt := ""
	for scanner.Scan() && count < 15 {
		count++
		excerpt = excerpt + scanner.Text() + " "
	}
	return excerpt
}

func CreatePost() {
	
}

func ReadPosts(res render.Render, db *r.Session) {
	var post Post
	posts, err := post.GetAll(db)
	if err != nil {
		res.JSON(500, err)
		return
	}
	res.JSON(200, posts)
}

func ReadPost(params martini.Params, res render.Render, db *r.Session) {
	var post Post
	post.Title = params["title"]
	post, err := post.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	res.JSON(200, post)
}

func (post Post) Get(s *r.Session) (Post, error) {
	row, err := r.Db(os.Getenv("rNAME")).Table("posts").Get(post.Title).RunRow(s)
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
