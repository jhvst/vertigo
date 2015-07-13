package routes

import (
	"log"
	"net/http"
	"time"

	. "github.com/9uuso/vertigo/databases/gorm"
	. "github.com/9uuso/vertigo/settings"

	"github.com/gorilla/feeds"
	"github.com/martini-contrib/render"
)

// ReadFeed renders RSS or Atom feed of latest published posts.
// It determines the feed type with strings.Split(r.URL.Path[1:], "/")[1].
func ReadFeed(w http.ResponseWriter, res render.Render, r *http.Request) {

	w.Header().Set("Content-Type", "application/xml")

	feed := &feeds.Feed{
		Title:       Settings.Name,
		Link:        &feeds.Link{Href: Settings.URL.String()},
		Description: Settings.Description,
	}

	var post Post
	posts, err := post.GetAll()
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	for _, post := range posts {

		var user User
		user.ID = post.Author
		user, err := user.Get()
		if err != nil {
			log.Println(err)
			res.JSON(500, map[string]interface{}{"error": "Internal server error"})
			return
		}

		// Don't expose unpublished items to the feeds
		if !post.Published {
			continue
		}

		// The email in &feeds.Author is not actually exported, as it is left out by user.Get().
		// However, the package panics if too few values are exported, so that will do.
		item := &feeds.Item{
			Title:       post.Title,
			Link:        &feeds.Link{Href: Settings.URL.String() + "/post/" + post.Slug},
			Description: post.Excerpt,
			Author:      &feeds.Author{user.Name, user.Email},
			Created:     time.Unix(post.Created, 0),
		}

		feed.Items = append(feed.Items, item)
	}

	result, err := feed.ToRss()
	if err != nil {
		log.Println(err)
		res.JSON(500, map[string]interface{}{"error": "Internal server error"})
		return
	}

	w.Write([]byte(result))
}
