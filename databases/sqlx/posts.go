package sqlx

import (
	"errors"
	"log"
	"time"

	"github.com/9uuso/excerpt"
	"github.com/9uuso/timezone"
	"github.com/martini-contrib/sessions"
	"github.com/russross/blackfriday"
	slug "github.com/shurcooL/sanitized_anchor_name"
)

// Post struct contains all relevant data when it comes to posts. Most fields
// are automatically filled when inserting new object into the database.
// JSON field after type refer to JSON key which martini will use to render data.
// Form field refers to frontend POST form `name` fields which martini uses to read data from.
// Binding defines whether the field is required when inserting or updating the object.
type Post struct {
	ID         int64  `json:"id"`
	Title      string `json:"title" form:"title" binding:"required"`
	Content    string `json:"content"`
	Markdown   string `json:"markdown" form:"markdown"`
	Slug       string `json:"slug"`
	Author     int64  `json:"author"`
	Excerpt    string `json:"excerpt"`
	Viewcount  uint   `json:"viewcount"`
	Published  bool   `json:"-"`
	Created    int64  `json:"created"`
	Updated    int64  `json:"updated"`
	TimeOffset int    `json:"timeoffset"`
}

// Insert or post.Insert inserts Post object into database.
// Requires active session cookie
// Fills post.Author, post.Created, post.Edited, post.Excerpt, post.Slug and post.Published automatically.
// Returns Post and error object.
func (post Post) Insert(s sessions.Session) (Post, error) {
	var user User
	user, err := user.Session(s)
	if err != nil {
		return post, err
	}
	_, offset, err := timezone.Offset(user.Location)
	if err != nil {
		return post, err
	}
	post.TimeOffset = offset
	post.Content = string(blackfriday.MarkdownCommon([]byte(post.Markdown)))
	post.Author = user.ID
	post.Created = time.Now().UTC().Round(time.Second).Unix()
	post.Updated = post.Created
	post.Excerpt = excerpt.Make(post.Content, 15)
	post.Slug = slug.Create(post.Title)
	post.Published = false
	post.Viewcount = 0
	_, err = db.Exec("INSERT INTO post (title, content, markdown, slug, author, excerpt, viewcount, published, created, updated, timeoffset) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		post.Title, post.Content, post.Markdown, post.Slug, user.ID, post.Excerpt, post.Viewcount, post.Published, post.Created, post.Updated, post.TimeOffset)
	if err != nil {
		return post, err
	}
	return post, nil
}

// Get or user.Get returns user according to given user.Slug.
// Requires session session as a parameter.
// Returns Ad and error object.
func (post Post) Get() (Post, error) {
	err := db.Get(&post, "SELECT * FROM post WHERE post.slug = ?", post.Slug)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return post, errors.New("not found")
		}
		return post, err
	}
	return post, nil
}

// Update or post.Update updates parameter "entry" with data given in parameter "post".
// Requires active session cookie.
// Returns updated Post object and an error object.
func (post Post) Update(s sessions.Session, entry Post) (Post, error) {
	var user User
	user, err := user.Session(s)
	if err != nil {
		return post, err
	}
	if post.Author == user.ID {
		entry.Content = string(blackfriday.MarkdownCommon([]byte(entry.Markdown)))
		entry.Excerpt = excerpt.Make(entry.Content, 15)
		entry.Slug = slug.Create(entry.Title)
		entry.Updated = time.Now().UTC().Round(time.Second).Unix()
		_, err := db.Exec(
			"UPDATE post SET title = ?, content = ?, markdown = ?, slug = ?, excerpt = ?, published = ?, updated = ? WHERE slug = ?",
			entry.Title, entry.Content, entry.Markdown, entry.Slug, entry.Excerpt, entry.Published, entry.Updated, post.Slug)
		if err != nil {
			return post, err
		}
		entry.ID = post.ID
		entry.Viewcount = post.Viewcount
		entry.Created = post.Created
		entry.TimeOffset = post.TimeOffset
		entry.Author = post.Author
		return entry, nil
	}
	return post, errors.New("unauthorized")
}

func (post Post) Unpublish(s sessions.Session) error {
	var user User
	user, err := user.Session(s)
	if err != nil {
		return err
	}
	if post.Author == user.ID {
		_, err := db.Exec("UPDATE post SET published = ? WHERE slug = ?", false, post.Slug)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("unauthorized")
}

// Delete or post.Delete deletes a post according to post.Slug.
// Requires session cookie.
// Returns error object.
func (post Post) Delete(s sessions.Session) error {
	var user User
	user, err := user.Session(s)
	if err != nil {
		return err
	}
	if post.Author == user.ID {
		_, err := db.Exec("DELETE FROM post WHERE slug = ?", post.Slug)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("unauthorized")
}

// GetAll or user.GetAll returns all user in database.
// Returns []User and error object.
func (post Post) GetAll() ([]Post, error) {
	var posts []Post
	err := db.Select(&posts, "SELECT * FROM post ORDER BY created")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			posts = make([]Post, 0)
			return posts, nil
		}
		return posts, err
	}
	return posts, nil
}

// Update or user.Update updates parameter "entry" with data given in parameter "user".
// Requires active session cookie.
// Returns updated Ad object and an error object.
func (post Post) Increment() {
	post.Viewcount += 1
	_, err := db.Exec("UPDATE post SET viewcount = ? WHERE slug = ?", post.Viewcount, post.Slug)
	if err != nil {
		log.Println("analytics error:", err)
	}
}
