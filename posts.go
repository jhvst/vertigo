// This file contains about everything related to posts. At the top you will find routes
// and at the bottom you can find CRUD options. Some functions in this file are analogous
// to the ones in users.go.
package main

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/9uuso/go-jaro-winkler-distance"
	r "github.com/dancannon/gorethink"
	"github.com/go-martini/martini"
	"github.com/gosimple/slug"
	"github.com/kennygrant/sanitize"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"log"
	"net/http"
	"strings"
	"time"
)

// Post struct contains all relevant data when it comes to posts. Most fields
// are automatically filled inserting a new object into the database.
// JSON field after type refer to JSON key which martini will use to render data.
// Form field refers to frontend POST form `name` fields which martini uses to read data from.
// Binding defines whether the field is required when inserting or updating the object.
// Gorethink field defines which name the variable gets once inserted to database.
type Post struct {
	Date      int64  `json:"date" gorethink:"date"`
	Title     string `json:"title" form:"title" binding:"required" gorethink:"title"`
	Author    string `json:"author,omitempty" gorethink:"author"`
	Content   string `json:",omitempty" form:"content" binding:"required" gorethink:"content"`
	Excerpt   string `json:"excerpt" gorethink:"excerpt"`
	Slug      string `json:"slug" gorethink:"slug"`
	Published bool   `json:"-" gorethink:"published"`
	Viewcount int32  `json:"viewcount" gorethink:"viewcount"`
	ID        string `json:"id" gorethink:",omitempty"`
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
func Homepage(res render.Render, db *r.Session) {
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
func SearchPost(req *http.Request, db *r.Session, res render.Render, search Search) {
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
func (search Search) Get(db *r.Session) ([]Post, error) {
	var matched []Post
	var post Post
	posts, err := post.GetAll(db)
	if err != nil {
		log.Println(err)
		return matched, err
	}
	for _, post := range posts {
		// posts are searched for a match in both content and title, so here
		// we declare two scanners for them
		content := bufio.NewScanner(strings.NewReader(post.Content))
		content.Split(bufio.ScanWords)
		title := bufio.NewScanner(strings.NewReader(post.Title))
		title.Split(bufio.ScanWords)
		// content is scanned trough Jaro-Winkler distance with
		// quite strict matching score of 0.9/1
		// matching score this high would most likely catch only different
		// capitalization and small, one letter missclicks in search query
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
	End:
	}
	return matched, nil
}

// CreatePost is a route which creates a new post according to the posted data.
// API response contains the created post object and normal request redirects to "/user" page.
// Does not publish the post automatically. See PublishPost for more.
func CreatePost(req *http.Request, s sessions.Session, db *r.Session, res render.Render, post Post) {
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
func ReadPosts(res render.Render, db *r.Session) {
	var post Post
	posts, err := post.GetAll(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	res.JSON(200, posts)
}

// ReadPost is a route which returns post with given post.Slug.
// Returns post data on JSON call and displays a formatted page on frontend.
func ReadPost(req *http.Request, s sessions.Session, params martini.Params, res render.Render, db *r.Session) {
	var post Post
	if params["title"] == "new" {
		res.JSON(406, map[string]interface{}{"error": "You cant name a post with colliding route name!"})
		return
	}
	post.Slug = params["title"]
	post, err := post.Get(db)
	go post.Increment(db)
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

// EditPost is a route which returns a post object to be displayed and edited on frontend.
// Not available for JSON API.
// Analogous to ReadPost. Could be replaced at some point.
func EditPost(req *http.Request, params martini.Params, res render.Render, db *r.Session) {
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
func UpdatePost(req *http.Request, params martini.Params, s sessions.Session, res render.Render, db *r.Session, post Post) {
	post.Slug = params["title"]
	entry, err := post.Get(db)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	post, err = entry.Update(db, s, post)
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
		res.Redirect("/user", 302)
		return
	}
	res.JSON(500, map[string]interface{}{"error": "Internal server error"})
}

// PublishPost is a route which publishes a post and therefore making it appear on frontpage and search.
// JSON request returns `HTTP 200 {"success": "Post published"}` on success. Frontend call will redirect to
// published page.
// Requires active session cookie.
func PublishPost(req *http.Request, params martini.Params, s sessions.Session, res render.Render, db *r.Session) {
	var post Post
	post.Slug = params["title"]
	post, err := post.Get(db)
	post.Published = true
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
		return
	}
	post, err = post.Update(db, s, post)
	if err != nil {
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
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
func DeletePost(req *http.Request, params martini.Params, s sessions.Session, res render.Render, db *r.Session) {
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
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		log.Println(err)
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
func (post Post) Insert(db *r.Session, s sessions.Session) (Post, error) {
	var person Person
	person, err := person.Session(db, s)
	if err != nil {
		log.Println(err)
		return post, err
	}
	post.Author = person.ID
	post.Date = time.Now().Unix()
	post.Excerpt = Excerpt(post.Content)
	post.Slug = slug.Make(post.Title)
	post.Published = false
	row, err := r.Table("posts").Insert(post).RunRow(db)
	if err != nil {
		log.Println(err)
		return post, err
	}
	err = row.Scan(&post)
	if err != nil {
		log.Println(err)
		return post, err
	}
	return post, err
}

// Get or post.Get returns post according to given post.Slug.
// Requires db session as a parameter.
// Returns Post and error object.
func (post Post) Get(db *r.Session) (Post, error) {
	row, err := r.Table("posts").Filter(func(this r.RqlTerm) r.RqlTerm {
		return this.Field("slug").Eq(post.Slug)
	}).RunRow(db)
	if err != nil {
		log.Println(err)
		return post, err
	}
	if row.IsNil() {
		return post, errors.New("nothing was found")
	}
	err = row.Scan(&post)
	if err != nil {
		log.Println(err)
		return post, err
	}
	return post, err
}

// Update or post.Update updates parameter "entry" with data given in parameter "post".
// Requires active session cookie.
// Returns updated Post object and an error object.
func (post Post) Update(db *r.Session, s sessions.Session, entry Post) (Post, error) {
	var person Person
	person, err := person.Session(db, s)
	if err != nil {
		log.Println(err)
		return post, err
	}
	if post.Author == person.ID {
		row, err := r.Table("posts").Filter(func(this r.RqlTerm) r.RqlTerm {
			return this.Field("slug").Eq(post.Slug)
		}).Update(map[string]interface{}{"published": entry.Published, "content": entry.Content, "slug": slug.Make(entry.Title), "title": entry.Title, "excerpt": Excerpt(entry.Content)}).RunRow(db)
		if err != nil {
			log.Println(err)
			return post, err
		}
		if row.IsNil() {
			return post, errors.New("nothing was found")
		}
		err = row.Scan(&post)
		if err != nil {
			log.Println(err)
			return post, err
		}
	} else {
		return post, errors.New("unauthorized")
	}
	return post, err
}

// Delete or post.Delete deletes a post according to post.Slug.
// Requires session cookie.
// Returns error object.
func (post Post) Delete(db *r.Session, s sessions.Session) error {
	var person Person
	person, err := person.Session(db, s)
	if err != nil {
		log.Println(err)
		return err
	}
	if post.Author == person.ID {
		row, err := r.Table("posts").Filter(func(this r.RqlTerm) r.RqlTerm {
			return this.Field("slug").Eq(post.Slug)
		}).Delete().RunRow(db)
		if err != nil {
			log.Println(err)
			return err
		}
		if row.IsNil() {
			return errors.New("nothing was found")
		}
		err = row.Scan(&post)
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		return errors.New("unauthorized")
	}
	return err
}

// GetAll or post.GetAll returns all posts in database.
// Returns []Post and error object.
func (post Post) GetAll(db *r.Session) ([]Post, error) {
	var posts []Post
	rows, err := r.Table("posts").OrderBy(r.Desc("date")).Run(db)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for rows.Next() {
		err := rows.Scan(&post)
		post, err := post.Get(db)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if post.Published {
			posts = append(posts, post)
		}
	}
	return posts, nil
}

// Increment or post.Increment increases viewcount of a post according to its post.ID
// It is supposed to be run as a gouroutine, so therefore it does not return anything.
func (post Post) Increment(db *r.Session) {
	_, err := r.Table("posts").Get(post.ID).Update(map[string]interface{}{"viewcount": post.Viewcount + 1}).RunRow(db)
	if err != nil {
		log.Println("analytics:", err)
	}
}