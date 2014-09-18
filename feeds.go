// Feeds.go contain HTTP routes for rendering RSS and Atom feeds.
package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	r "github.com/dancannon/gorethink"
	"github.com/gorilla/feeds"
	"github.com/martini-contrib/render"
)

// ReadFeed renders RSS or Atom feed of latest published posts.
// It determines the feed type with strings.Split(r.URL.Path[1:], "/")[1].
func ReadFeed(w http.ResponseWriter, res render.Render, db *r.Session, r *http.Request) {

	w.Header().Set("Content-Type", "application/xml")

	feed := &feeds.Feed{
		Title:       Settings.Hostname,
		Link:        &feeds.Link{Href: "http://" + Settings.Hostname},
		Description: Settings.Description,
	}

	var post Post
	posts, err := post.GetAll(db)
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	for _, post := range posts {

		var person Person
		person.ID = post.Author
		person, err := person.Get(db)
		if err != nil {
			log.Println(err)
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
		}

		// The email in &feeds.Author is not actually exported, as it is left out by person.Get().
		// However, the package panics if too few values are exported, so that will do.
		item := &feeds.Item{
			Title:       post.Title,
			Link:        &feeds.Link{Href: "http://" + Settings.Hostname + "/" + post.Slug},
			Description: post.Excerpt,
			Author:      &feeds.Author{person.Name, person.Email},
			Created:     time.Unix(post.Date, 0),
		}
		feed.Items = append(feed.Items, item)

	}

	// Default to RSS feed.
	result, err := feed.ToRss()
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	format := strings.Split(r.URL.Path[1:], "/")[1]
	if format == "atom" {
		result, err = feed.ToAtom()
		if err != nil {
			log.Println(err)
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
		}
	}

	w.Write([]byte(result))
}
