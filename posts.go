// This file contains about everything related to posts. At the top you will find routes
// and at the bottom you can find CRUD options. Some functions in this file are analogous
// to the ones in users.go.
package main

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/russross/blackfriday"
	"github.com/9uuso/go-jaro-winkler-distance"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/kennygrant/sanitize"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	_ "github.com/mattn/go-sqlite3"
)

// Post struct contains all relevant data when it comes to posts. Most fields
// are automatically filled when inserting new object into the database.
// JSON field after type refer to JSON key which martini will use to render data.
// Form field refers to frontend POST form `name` fields which martini uses to read data from.
// Binding defines whether the field is required when inserting or updating the object.
type Post struct {
	ID        int64  `json:"id"`
	Title     string `json:"title" form:"title" binding:"required"`
	Content   string `json:"content" form:"content" binding:"required"`
	Date      int64  `json:"date"`
	Slug      string `json:"slug"`
	Author    int64  `json:"author"`
	Excerpt   string `json:"excerpt"`
	Viewcount uint   `json:"viewcount"`
	Published bool   `json:"-"`
}

// Search struct is basically just a type check to make sure people don't add anything nasty to
// on-site search queries.
type Search struct {
	Query string `json:"query" form:"query" binding:"required"`
	Score float64
	Post  Post
}

// Homepage route fetches all posts from database and renders them according to "home.tmpl".
// Normally you'd use this function as your "/" route.
func Homepage(res render.Render, db *gorm.DB) {
	if Settings.Firstrun {
		res.HTML(200, "installation/wizard", nil)
		return
	}
	var post Post
	posts, err := post.GetAll(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	res.HTML(200, "home", posts)
}

// Excerpt generates 15 word excerpt from given input.
// Used to make shorter summaries from blog posts.
func Excerpt(input string) string {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)
	count := 0
	var excerpt bytes.Buffer
	for scanner.Scan() && count < 15 {
		count++
		excerpt.WriteString(scanner.Text() + " ")
	}
	return sanitize.HTML(strings.TrimSpace(excerpt.String()))
}

// SearchPost is a route which returns all posts and aggregates the ones which contain
// the POSTed search query in either Title or Content field.
func SearchPost(req *http.Request, db *gorm.DB, res render.Render, search Search) {
	posts, err := search.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	switch root(req) {
	case "api":
		res.JSON(200, posts)
		return
	case "post":
		res.HTML(200, "search", posts)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// Get or search.Get returns all posts which contain parameter search.Query in either
// post.Title or post.Content.
// Returns []Post and error object.
func (search Search) Get(db *gorm.DB) ([]Post, error) {
	var matched []Post
	var post Post
	posts, err := post.GetAll(db)
	if err != nil {
		log.Println(err)
		return matched, err
	}
	for _, post := range posts {
		if post.Published {
			// posts are searched for a match in both content and title, so here
			// we declare two scanners for them
			content := bufio.NewScanner(strings.NewReader(post.Content))
			content.Split(bufio.ScanWords)
			title := bufio.NewScanner(strings.NewReader(post.Title))
			title.Split(bufio.ScanWords)
			// content is scanned trough Jaro-Winkler distance with
			// quite strict matching score of 0.9/1
			// matching score this high would most likely catch only different
			// capitalization and small typos
			//
			// since we are already in a for loop, we have to break the
			// iteration here by going to label End to avoid showing a
			// duplicate search result
			for content.Scan() {
				if jwd.Calculate(content.Text(), search.Query) >= 0.9 {
					matched = append(matched, post)
					goto End
				}
			}
			for title.Scan() {
				if jwd.Calculate(title.Text(), search.Query) >= 0.9 {
					matched = append(matched, post)
					goto End
				}
			}
		}
	End:
	}
	return matched, nil
}

// CreatePost is a route which creates a new post according to the posted data.
// API response contains the created post object and normal request redirects to "/user" page.
// Does not publish the post automatically. See PublishPost for more.
func CreatePost(req *http.Request, s sessions.Session, db *gorm.DB, res render.Render, post Post) {
	entry, err := post.Insert(db, s)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	switch root(req) {
	case "api":
		res.JSON(200, entry)
		return
	case "post":
		res.Redirect("/user", 302)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// ReadPosts is a route which returns all posts without merged owner data (although the object does include author field)
// Not available on frontend, so therefore it only returns a JSON response.
func ReadPosts(res render.Render, db *gorm.DB) {
	var post Post
	var published []Post
	posts, err := post.GetAll(db)
	for _, post := range posts {
		if post.Published {
			published = append(published, post)
		}
	}
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	res.JSON(200, published)
}

// ReadPost is a route which returns post with given post.Slug.
// Returns post data on JSON call and displays a formatted page on frontend.
func ReadPost(req *http.Request, s sessions.Session, params martini.Params, res render.Render, db *gorm.DB) {
	var post Post
	if params["title"] == "new" {
		res.JSON(406, map[string]interface{}{"error": "You cant name a post with colliding route name!"})
		return
	}
	post.Slug = params["title"]
	post, err := post.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	if post.Published {
		go post.Increment(db)
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

// EditPost is a route which returns a post object to be displayed and edited on frontend.
// Not available for JSON API.
// Analogous to ReadPost. Could be replaced at some point.
func EditPost(req *http.Request, params martini.Params, res render.Render, db *gorm.DB) {
	var post Post
	post.Slug = params["title"]
	post, err := post.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	switch root(req) {
	case "api":
		res.JSON(403, map[string]interface{}{"error": "To edit a post POST to /api/post/:title/edit instead."})
		return
	case "post":
		res.HTML(200, "post/edit", post)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// UpdatePost is a route which updates a post defined by martini parameter "title" with posted data.
// Requires session cookie. JSON request returns the updated post object, frontend call will redirect to "/user".
func UpdatePost(req *http.Request, params martini.Params, s sessions.Session, res render.Render, db *gorm.DB, post Post) {
	post.Slug = params["title"]
	entry, err := post.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	post, err = entry.Update(db, s, post)
	if err != nil {
		if err.Error() == "unauthorized" {
			res.JSON(401, map[string]interface{}{"error": "You are not allowed to do that"})
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch root(req) {
	case "api":
		res.JSON(200, post)
		return
	case "post":
		res.Redirect("/user", 302)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// PublishPost is a route which publishes a post and therefore making it appear on frontpage and search.
// JSON request returns `HTTP 200 {"success": "Post published"}` on success. Frontend call will redirect to
// published page.
// Requires active session cookie.
func PublishPost(req *http.Request, params martini.Params, s sessions.Session, res render.Render, db *gorm.DB) {
	var post Post
	post.Slug = params["title"]
	post.Published = true
	old, err := post.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	post, err = old.Update(db, s, post)
	if err != nil {
		if err.Error() == "unauthorized" {
			res.JSON(401, map[string]interface{}{"error": "Unauthorized"})
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch root(req) {
	case "api":
		res.JSON(200, map[string]interface{}{"success": "Post published"})
		return
	case "post":
		res.Redirect("/post/"+post.Slug, 302)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// DeletePost is a route which deletes a post according to martini parameter "title".
// JSON request returns `HTTP 200 {"success": "Post deleted"}` on success. Frontend call will redirect to
// "/user" page on successful request.
// Requires active session cookie.
func DeletePost(req *http.Request, params martini.Params, s sessions.Session, res render.Render, db *gorm.DB) {
	var post Post
	post.Slug = params["title"]
	post, err := post.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	err = post.Delete(db, s)
	if err != nil {
		if err.Error() == "unauthorized" {
			res.JSON(401, map[string]interface{}{"error": "Unauthorized"})
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch root(req) {
	case "api":
		res.JSON(200, map[string]interface{}{"success": "Post deleted"})
		return
	case "post":
		res.Redirect("/user", 302)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// Insert or post.Insert inserts Post object into database.
// Requires active session cookie
// Fills post.Author, post.Date, post.Excerpt, post.Slug and post.Published automatically.
// Returns Post and error object.
func (post Post) Insert(db *gorm.DB, s sessions.Session) (Post, error) {
	var user User
	user, err := user.Session(db, s)
	if err != nil {
		log.Println(err)
		return post, err
	}
	post.Content = string(blackfriday.MarkdownCommon([]byte(post.Content)))
	post.Author = user.ID
	post.Date = time.Now().Unix()
	post.Excerpt = Excerpt(post.Content)
	post.Slug = slug.Make(post.Title)
	post.Published = false
	db.Create(&post)
	if db.Error != nil {
		log.Println(db.Error)
		return post, db.Error
	}
	return post, nil
}

// Get or post.Get returns post according to given post.Slug.
// Requires db session as a parameter.
// Returns Post and error object.
func (post Post) Get(db *gorm.DB) (Post, error) {
	db.Find(&post, Post{Slug: post.Slug})
	if db.Error != nil {
		log.Println(db.Error)
		return post, db.Error
	}
	return post, nil
}

// Update or post.Update updates parameter "entry" with data given in parameter "post".
// Requires active session cookie.
// Returns updated Post object and an error object.
func (post Post) Update(db *gorm.DB, s sessions.Session, entry Post) (Post, error) {
	var user User
	user, err := user.Session(db, s)
	if err != nil {
		log.Println(err)
		return post, err
	}
	if post.Author == user.ID {
		entry.Content = string(blackfriday.MarkdownCommon([]byte(entry.Content)))		
		db.Where(&Post{Slug: post.Slug}).Find(&post).Updates(entry)
		if db.Error != nil {
			log.Println(db.Error)
			return post, db.Error
		}
		return post, nil
	}
	return post, errors.New("unauthorized")
}

// Delete or post.Delete deletes a post according to post.Slug.
// Requires session cookie.
// Returns error object.
func (post Post) Delete(db *gorm.DB, s sessions.Session) error {
	var user User
	user, err := user.Session(db, s)
	if err != nil {
		log.Println(err)
		return err
	}
	if post.Author == user.ID {
		db.Where(&Post{Slug: post.Slug}).Delete(&post)
		if db.Error != nil {
			log.Println(db.Error)
			return db.Error
		}
	} else {
		return errors.New("unauthorized")
	}
	return nil
}

// GetAll or post.GetAll returns all posts in database.
// Returns []Post and error object.
func (post Post) GetAll(db *gorm.DB) ([]Post, error) {
	var posts []Post
	db.Find(&posts)
	if db.Error != nil {
		log.Println(db.Error)
		return posts, db.Error
	}
	return posts, nil
}

// Increment or post.Increment increases viewcount of a post according to its post.ID
// It is supposed to be run as a gouroutine, so therefore it does not return anything.
func (post Post) Increment(db *gorm.DB) {
	var entry Post
	entry.Viewcount = post.Viewcount + 1
	db.Where(&Post{Slug: post.Slug}).Find(&post).Updates(entry)
	if db.Error != nil {
		log.Println("analytics", db.Error)
	}
}
