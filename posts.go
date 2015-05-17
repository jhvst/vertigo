// This file contains about everything related to posts. At the top you will find routes
// and at the bottom you can find CRUD options. Some functions in this file are analogous
// to the ones in users.go.
package main

import (
	"bufio"
	"log"
	"net/http"
	"strings"

	. "github.com/9uuso/vertigo/databases/gorm"
	. "github.com/9uuso/vertigo/misc"
	. "github.com/9uuso/vertigo/settings"

	"github.com/9uuso/go-jaro-winkler-distance"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
)

// Homepage route fetches all posts from database and renders them according to "home.tmpl".
// Normally you'd use this function as your "/" route.
func Homepage(res render.Render) {
	if Settings.Firstrun {
		res.HTML(200, "installation/wizard", nil)
		return
	}
	var post Post
	posts, err := post.GetAll()
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	res.HTML(200, "home", posts)
}

// Search struct is basically just a type check to make sure people don't add anything nasty to
// on-site search queries.
type Search struct {
	Query string `json:"query" form:"query" binding:"required"`
	Score float64
	Posts []Post
}

// Get or search.Get returns all posts which contain parameter search.Query in either
// post.Title or post.Content.
// Returns []Post and error object.
func (search Search) Get() (Search, error) {
	var post Post
	posts, err := post.GetAll()
	if err != nil {
		log.Println(err)
		return search, err
	}
	for _, post := range posts {
		if post.Published {
			// posts are searched for a match in both content and title, so here
			// we declare two scanners for them
			content := bufio.NewScanner(strings.NewReader(post.Content))
			title := bufio.NewScanner(strings.NewReader(post.Title))
			// Blackfriday makes smartypants corrections some characters, which break the search
			if Settings.Markdown {
				content = bufio.NewScanner(strings.NewReader(post.Markdown))
				title = bufio.NewScanner(strings.NewReader(post.Title))
			}
			content.Split(bufio.ScanWords)
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
					search.Posts = append(search.Posts, post)
					goto End
				}
			}
			for title.Scan() {
				if jwd.Calculate(title.Text(), search.Query) >= 0.9 {
					search.Posts = append(search.Posts, post)
					goto End
				}
			}
		}
	End:
	}
	if len(search.Posts) == 0 {
		search.Posts = make([]Post, 0)
	}
	return search, nil
}

// SearchPost is a route which returns all posts and aggregates the ones which contain
// the POSTed search query in either Title or Content field.
func SearchPost(req *http.Request, res render.Render, search Search) {
	search, err := search.Get()
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch Root(req) {
	case "api":
		res.JSON(200, search.Posts)
		return
	case "post":
		res.HTML(200, "search", search.Posts)
		return
	}
}

// CreatePost is a route which creates a new post according to the posted data.
// API response contains the created post object and normal request redirects to "/user" page.
// Does not publish the post automatically. See PublishPost for more.
func CreatePost(req *http.Request, s sessions.Session, res render.Render, post Post) {
	post, err := post.Insert(s)
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch Root(req) {
	case "api":
		res.JSON(200, post)
		return
	case "post":
		res.Redirect("/user", 302)
		return
	}
}

// ReadPosts is a route which returns all posts without merged owner data (although the object does include author field)
// Not available on frontend, so therefore it only returns a JSON response.
func ReadPosts(res render.Render) {
	var post Post
	published := make([]Post, 0)
	posts, err := post.GetAll()
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	for _, post := range posts {
		if post.Published {
			published = append(published, post)
		}
	}
	res.JSON(200, published)
}

// ReadPost is a route which returns post with given post.Slug.
// Returns post data on JSON call and displays a formatted page on frontend.
func ReadPost(req *http.Request, s sessions.Session, params martini.Params, res render.Render) {
	var post Post
	if params["slug"] == "new" {
		res.JSON(400, map[string]interface{}{"error": "There can't be a post called 'new'."})
		return
	}
	post.Slug = params["slug"]
	post, err := post.Get()
	if err != nil {
		log.Println(err)
		if err.Error() == "not found" {
			res.JSON(404, NotFound())
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	go post.Increment()
	switch Root(req) {
	case "api":
		res.JSON(200, post)
		return
	case "post":
		res.HTML(200, "post/display", post)
		return
	}
}

// EditPost is a route which returns a post object to be displayed and edited on frontend.
// Not available for JSON API.
// Analogous to ReadPost. Could be replaced at some point.
func EditPost(req *http.Request, params martini.Params, res render.Render) {
	var post Post
	post.Slug = params["slug"]
	post, err := post.Get()
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	res.HTML(200, "post/edit", post)
}

// UpdatePost is a route which updates a post defined by martini parameter "title" with posted data.
// Requires session cookie. JSON request returns the updated post object, frontend call will redirect to "/user".
func UpdatePost(req *http.Request, params martini.Params, s sessions.Session, res render.Render, entry Post) {
	var post Post
	post.Slug = params["slug"]
	post, err := post.Get()
	if err != nil {
		log.Println(err)
		if err.Error() == "not found" {
			res.JSON(404, NotFound())
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	var user User
	user, err = user.Session(s)
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	if post.Author == user.ID {
		post, err = post.Update(entry)
		if err != nil {
			log.Println(err)
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
		}
	} else {
		res.JSON(401, map[string]interface{}{"error": "Unauthorized"})
		return
	}

	switch Root(req) {
	case "api":
		res.JSON(200, post)
		return
	case "post":
		res.Redirect("/user", 302)
		return
	}
}

// PublishPost is a route which publishes a post and therefore making it appear on frontpage and search.
// JSON request returns `HTTP 200 {"success": "Post published"}` on success. Frontend call will redirect to
// published page.
// Requires active session cookie.
func PublishPost(req *http.Request, params martini.Params, s sessions.Session, res render.Render) {
	var post Post
	post.Slug = params["slug"]
	post, err := post.Get()
	if err != nil {
		if err.Error() == "not found" {
			res.JSON(404, NotFound())
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	var user User
	user, err = user.Session(s)
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	if post.Author == user.ID {
		var entry Post
		entry.Published = true
		post, err = post.Update(entry)
		if err != nil {
			log.Println(err)
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
		}
	} else {
		res.JSON(401, map[string]interface{}{"error": "Unauthorized"})
		return
	}

	switch Root(req) {
	case "api":
		res.JSON(200, map[string]interface{}{"success": "Post published"})
		return
	case "post":
		res.Redirect("/post/"+post.Slug, 302)
		return
	}
}

// UnpublishPost is a route which unpublishes a post and therefore making it disappear from frontpage and search.
// JSON request returns `HTTP 200 {"success": "Post unpublished"}` on success. Frontend call will redirect to
// user control panel.
// Requires active session cookie.
// The route is anecdotal to route PublishPost().
func UnpublishPost(req *http.Request, params martini.Params, s sessions.Session, res render.Render) {
	var post Post
	post.Slug = params["slug"]
	post, err := post.Get()
	if err != nil {
		if err.Error() == "not found" {
			res.JSON(404, NotFound())
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	var user User
	user, err = user.Session(s)
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	if post.Author == user.ID {
		err = post.Unpublish(s)
		if err != nil {
			log.Println(err)
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
		}
	} else {
		res.JSON(401, map[string]interface{}{"error": "Unauthorized"})
		return
	}

	switch Root(req) {
	case "api":
		res.JSON(200, map[string]interface{}{"success": "Post unpublished"})
		return
	case "post":
		res.Redirect("/user", 302)
		return
	}
}

// DeletePost is a route which deletes a post according to martini parameter "title".
// JSON request returns `HTTP 200 {"success": "Post deleted"}` on success. Frontend call will redirect to
// "/user" page on successful request.
// Requires active session cookie.
func DeletePost(req *http.Request, params martini.Params, s sessions.Session, res render.Render) {
	var post Post
	post.Slug = params["slug"]
	post, err := post.Get()
	if err != nil {
		if err.Error() == "not found" {
			res.JSON(404, NotFound())
			return
		}
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	err = post.Delete(s)
	if err != nil {
		log.Println(err)
		if err.Error() == "unauthorized" {
			res.JSON(401, map[string]interface{}{"error": "Unauthorized"})
			return
		}
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}
	switch Root(req) {
	case "api":
		res.JSON(200, map[string]interface{}{"success": "Post deleted"})
		return
	case "post":
		res.Redirect("/user", 302)
		return
	}
}
