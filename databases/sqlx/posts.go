package sqlx

import (
	"errors"
	"log"
	"time"

	"github.com/9uuso/excerpt"
	"github.com/9uuso/timezone"
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
func (post Post) Insert(user User) (Post, error) {
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
	_, err = db.NamedExec(`INSERT INTO posts (title, content, markdown, slug, author, excerpt, viewcount, published, created, updated, timeoffset)
		VALUES (:title, :content, :markdown, :slug, :author, :excerpt, :viewcount, :published, :created, :updated, :timeoffset)`, post)
	if err != nil {
		return post, err
	}
	return post, nil
}

// Get or user.Get returns user according to given user.Slug.
// Requires session session as a parameter.
// Returns Ad and error object.
func (post Post) Get() (Post, error) {
	stmt, err := db.PrepareNamed("SELECT * FROM posts WHERE slug = :slug")
	if err != nil {
		return post, err
	}
	err = stmt.Get(&post, post)
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
func (post Post) Update(entry Post) (Post, error) {
	entry.ID = post.ID
	entry.Content = string(blackfriday.MarkdownCommon([]byte(entry.Markdown)))
	entry.Excerpt = excerpt.Make(entry.Content, 15)
	entry.Slug = slug.Create(entry.Title)
	entry.Updated = time.Now().UTC().Round(time.Second).Unix()
	_, err := db.NamedExec(
		"UPDATE posts SET title = :title, content = :content, markdown = :markdown, slug = :slug, excerpt = :excerpt, published = :published, updated = :updated WHERE id = :id",
		entry)
	if err != nil {
		return post, err
	}
	entry.Viewcount = post.Viewcount
	entry.Created = post.Created
	entry.TimeOffset = post.TimeOffset
	entry.Author = post.Author
	return entry, nil
}

func (post Post) Unpublish() error {
	post.Published = false
	_, err := db.NamedExec("UPDATE posts SET published = :published WHERE id = :id", post)
	if err != nil {
		return err
	}
	return nil
}

// Delete or post.Delete deletes a post according to post.Slug.
// Requires session cookie.
// Returns error object.
func (post Post) Delete() error {
	_, err := db.NamedExec("DELETE FROM posts WHERE id = :id", post)
	if err != nil {
		return err
	}
	return nil
}

// GetAll or user.GetAll returns all user in database.
// Returns []User and error object.
func (post Post) GetAll() ([]Post, error) {
	var posts []Post
	rows, err := db.Queryx("SELECT * FROM posts ORDER BY created DESC")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			posts = make([]Post, 0)
			return posts, nil
		}
		return posts, err
	}
	for rows.Next() {
		err := rows.StructScan(&post)
		if err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// Update or user.Update updates parameter "entry" with data given in parameter "user".
// Returns updated Ad object and an error object.
func (post Post) Increment() {
	post.Viewcount += 1
	_, err := db.NamedExec("UPDATE posts SET viewcount = :viewcount WHERE id = :id", post)
	if err != nil {
		log.Println("analytics error:", err)
	}
}
