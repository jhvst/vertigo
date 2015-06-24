package gorm

import (
	"errors"
	"log"
	"time"

	. "github.com/9uuso/vertigo/misc"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/sessions"
	"github.com/russross/blackfriday"
)

// Post struct contains all relevant data when it comes to posts. Most fields
// are automatically filled when inserting new object into the database.
// JSON field after type refer to JSON key which martini will use to render data.
// Form field refers to frontend POST form `name` fields which martini uses to read data from.
// Binding defines whether the field is required when inserting or updating the object.
type Post struct {
	ID        int64  `json:"id" gorm:"primary_key:yes"`
	Title     string `json:"title" form:"title" binding:"required"`
	Content   string `json:"content" form:"content" sql:"type:text"`
	Markdown  string `json:"markdown" form:"markdown" sql:"type:text"`
	Slug      string `json:"slug"`
	Author    int64  `json:"author"`
	Excerpt   string `json:"excerpt"`
	Viewcount uint   `json:"viewcount"`
	Published bool   `json:"-"`
	Created   int64  `json:"created"`
	Updated   int64  `json:"updated"`
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
	// if post.Content is empty, the user has used Markdown editor
	post.Markdown = Cleanup(post.Markdown)
	post.Content = string(blackfriday.MarkdownCommon([]byte(post.Markdown)))	
	post.Author = user.ID
	post.Created = time.Now().Unix()
	post.Updated = post.Created
	post.Excerpt = Excerpt(post.Content)
	post.Slug = slug.Make(post.Title)
	post.Published = false
	query := connection.Gorm.Create(&post)
	if query.Error != nil {
		return post, query.Error
	}
	return post, nil
}

// Get or post.Get returns post according to given post.Slug.
// Returns Post and error object.
func (post Post) Get() (Post, error) {
	query := connection.Gorm.Find(&post, Post{Slug: post.Slug})
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			return post, errors.New("not found")
		}
		return post, query.Error
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
		entry.Markdown = Cleanup(entry.Markdown)
		entry.Content = string(blackfriday.MarkdownCommon([]byte(entry.Markdown)))
		// this closure would need a call to convert HTML to Markdown
		// see https://github.com/9uuso/vertigo/issues/7
		// entry.Markdown = Markdown of entry.Content
		entry.Content = Cleanup(entry.Content)
		entry.Excerpt = Excerpt(entry.Content)
		entry.Slug = slug.Make(entry.Title)
		entry.Updated = time.Now().Unix()
		query := connection.Gorm.Where(&Post{Slug: post.Slug, Author: user.ID}).First(&post).Updates(entry)
		if query.Error != nil {
			if query.Error == gorm.RecordNotFound {
				return post, errors.New("not found")
			}
			return post, query.Error
		}
		return post, nil
	}
	return post, errors.New("unauthorized")
}

// Unpublish or post.Unpublish unpublishes a post by updating the Published value to false.
// Gorm specific, declared only because the library is buggy.
func (post Post) Unpublish(s sessions.Session) error {
	var user User
	user, err := user.Session(s)
	if err != nil {
		return err
	}
	if post.Author == user.ID {
		query := connection.Gorm.Where(&Post{Slug: post.Slug}).Find(&post).Update("published", false)
		if query.Error != nil {
			if query.Error == gorm.RecordNotFound {
				return errors.New("not found")
			}
			return query.Error
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
		query := connection.Gorm.Where(&Post{Slug: post.Slug}).Delete(&post)
		if query.Error != nil {
			if query.Error == gorm.RecordNotFound {
				return errors.New("not found")
			}
			return query.Error
		}
	} else {
		return errors.New("unauthorized")
	}
	return nil
}

// GetAll or post.GetAll returns all posts in database.
// Returns []Post and error object.
func (post Post) GetAll() ([]Post, error) {
	var posts []Post
	query := connection.Gorm.Order("created desc").Find(&posts)
	if query.Error != nil {
		if query.Error == gorm.RecordNotFound {
			posts = make([]Post, 0)
			return posts, nil
		}
		return posts, query.Error
	}
	return posts, nil
}

// Increment or post.Increment increases viewcount of a post according to its post.ID
// It is supposed to be run as a gouroutine, so therefore it does not return anything.
func (post Post) Increment() {
	var entry Post
	post.Viewcount += 1
	query := connection.Gorm.Where(&Post{Slug: post.Slug}).First(&entry).Update("viewcount", post.Viewcount)
	if query.Error != nil {
		log.Println("analytics error:", query.Error)
	}
}
