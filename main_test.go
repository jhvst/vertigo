package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/PuerkitoBio/goquery"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var sessioncookie *string = flag.String("sessioncookie", "", "global flag for test sessioncookie")
var postslug *string = flag.String("postslug", "", "global flag for test postslug")

var _ = Describe("Vertigo", func() {

	var server Server
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	BeforeEach(func() {
		// Set up a new server, connected to a test database,
		// before each test.
		server = NewServer()

		// Record HTTP responses.
		recorder = httptest.NewRecorder()
	})

	Describe("Web server and installation wizard", func() {

		// Set up a new GET request before every test
		// in this describe block.
		BeforeEach(func() {
			request, _ = http.NewRequest("GET", "/", nil)
		})

		Context("loading the homepage", func() {
			It("returns a status code of 200", func() {
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("first link's value should be 'Home'", func() {
				server.ServeHTTP(recorder, request)

				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}

				sel := doc.Find("a").First().Text()
				Expect(sel).To(Equal("Home"))
			})

			It("page's should display installation wizard", func() {
				server.ServeHTTP(recorder, request)

				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}

				sel := doc.Find("h1").First().Text()
				Expect(sel).To(Equal("Your settings file seems to miss some fields. Lets fix that."))
			})
		})
	})

	Describe("Settings", func() {

		Context("after creation", func() {

			It("Firstrun should equal to true", func() {
				settings := VertigoSettings()
				Expect(settings.Firstrun).To(Equal(true))
			})

		})

		Context("after submitting settings in JSON", func() {

			It("response should be a redirection", func() {
				request, err := http.NewRequest("POST", "/api/installation", strings.NewReader(`{"hostname": "example.com", "name": "Foo Blog", "description": "Foo's test blog", "mailgun": {"mgdomain": "foo", "mgprikey": "foo", "mgpubkey": "foo"}}`))
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("the settings.json should have all fields populated", func() {
				Expect(Settings.Hostname).To(Equal("example.com"))
				Expect(Settings.Name).To(Equal("Foo Blog"))
				Expect(Settings.Description).To(Equal("Foo's test blog"))
				Expect(Settings.Mailer.Domain).To(Equal("foo"))
				Expect(Settings.Mailer.PrivateKey).To(Equal("foo"))
				Expect(Settings.Mailer.PrivateKey).To(Equal("foo"))
				Expect(Settings.Mailer.PublicKey).To(Equal("foo"))
			})

		})

		Context("when manipulating the global Settings variable", func() {

			It("should save the changes to disk", func() {
				var settings Vertigo
				settings.Name = "Juuso's Blog"
				err := settings.Save()
				if err != nil {
					panic(err)
				}
				Expect(Settings.Name).To(Equal("Juuso's Blog"))
			})

			It("frontpage's <title> should now be 'Juuso's Blog'", func() {
				request, err := http.NewRequest("GET", "/", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}
				sel := doc.Find("title").Text()
				Expect(sel).To(Equal("Juuso's Blog"))
			})
		})
	})

	Describe("Users", func() {

		Context("creation", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("POST", "/api/user", strings.NewReader(`{"name": "Juuso", "password": "foo", "email": "foo@example.com"}`))
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

		})

		Context("reading", func() {

			It("should be then listed on /users", func() {
				request, err := http.NewRequest("GET", "/api/users", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				var users []User
				if err := json.Unmarshal(recorder.Body.Bytes(), &users); err != nil {
					panic(err)
				}
				Expect(recorder.Code).To(Equal(200))
				fmt.Println(recorder.Body)
				for i, user := range users {
					Expect(i).To(Equal(0))
					Expect(user.Name).To(Equal("Juuso"))
					Expect(user.ID).NotTo(Equal(""))
				}
			})
		})

		Context("accessing control panel before signing", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("GET", "/user", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(302))
			})
		})

		Context("signing in", func() {

			It("should return HTTP 200", func() {

				request, err := http.NewRequest("POST", "/api/user/login", strings.NewReader(`{"name": "Juuso", "password": "foo", "email": "foo@example.com"}`))
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				// i assure, nothing else worked
				cookie := strings.Split(strings.TrimLeft(recorder.HeaderMap["Set-Cookie"][0], "user="), ";")[0]
				flag.Set("sessioncookie", cookie)
				fmt.Println("User sessioncookie:", *sessioncookie)
				Expect(recorder.Code).To(Equal(200))

			})
		})

		Context("accessing control panel after signing", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("GET", "/user", nil)
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})
		})

	})

	Describe("Posts", func() {

		Context("creation", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("POST", "/api/post", strings.NewReader(`{"title": "Example post", "content": "This is example post with HTML elements like <b>bold</b> and <i>italics</i> in place."}`))
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				var post Post
				if err := json.Unmarshal(recorder.Body.Bytes(), &post); err != nil {
					panic(err)
				}
				flag.Set("postslug", post.Slug)
			})
		})

		Context("reading", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("GET", "/api/post/"+*postslug, nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})
		})

		Context("publishing", func() {

			It("without session data should return HTTP 401", func() {
				request, err := http.NewRequest("GET", "/api/post/"+*postslug+"/publish", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
			})

			It("with session data should return HTTP 200", func() {
				request, err := http.NewRequest("GET", "/api/post/"+*postslug+"/publish", nil)
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("after publishing, the post should be displayed on frontpage", func() {
				request, err := http.NewRequest("GET", "/", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}
				sel := doc.Find("article h1").Text()
				Expect(sel).To(Equal("Example post"))
			})
		})

		Context("reading after publishing", func() {

			It("should display the new post", func() {
				request, err := http.NewRequest("GET", "/api/posts", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body).NotTo(Equal("null"))
				var posts []Post
				if err := json.Unmarshal(recorder.Body.Bytes(), &posts); err != nil {
					panic(err)
				}
				for i, post := range posts {
					Expect(i).To(Equal(0))
					Expect(post.Slug).To(Equal(*postslug))
					Expect(post.Title).To(Equal("Example post"))
					Expect(post.Viewcount).To(Equal(uint(0)))
					Expect(post.Excerpt).To(Equal("This is example post with HTML elements like bold and italics in place."))
					Expect(post.Content).To(Equal("This is example post with HTML elements like <b>bold</b> and <i>italics</i> in place."))
				}
			})
		})

		Context("updating", func() {

			It("should return the updated post structure", func() {
				request, err := http.NewRequest("POST", "/api/post/"+*postslug+"/edit", strings.NewReader(`{"title": "Example post edited", "content": "This is an EDITED example post with HTML elements like <b>bold</b> and <i>italics</i> in place."}`))
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				var post Post
				if err := json.Unmarshal(recorder.Body.Bytes(), &post); err != nil {
					panic(err)
				}
				Expect(post.Title).To(Equal("Example post edited"))
				flag.Set("postslug", post.Slug)
			})
		})

		Context("reading after updating", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("GET", "/api/post/"+*postslug, nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})
		})
	})

})
