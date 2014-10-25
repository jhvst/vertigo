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
var globalpost *Post = &Post{}

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
				request, err := http.NewRequest("POST", "/api/installation", strings.NewReader(`{"hostname": "example.com", "name": "Foo Blog", "description": "Foo's test blog", "mailgun": {"mgdomain": "foo", "mgprikey": "foo"}}`))
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("the settings.json should have all fields populated", func() {
				Expect(Settings.Hostname).To(Equal("example.com"))
				Expect(Settings.AllowRegistrations).To(Equal(true))
				Expect(Settings.Markdown).To(Equal(false))
				Expect(Settings.Name).To(Equal("Foo Blog"))
				Expect(Settings.Description).To(Equal("Foo's test blog"))
				Expect(Settings.Mailer.Domain).To(Equal("foo"))
				Expect(Settings.Mailer.PrivateKey).To(Equal("foo"))
			})

		})

		Context("when manipulating the global Settings variable", func() {

			It("should save the changes to disk", func() {
				var settings *Vertigo
				settings = Settings
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
				payload := `{"name": "Juuso", "password": "foo", "email": "foo@example.com"}`
				request, err := http.NewRequest("POST", "/api/user", strings.NewReader(payload))
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				Expect(recorder.Body.String()).To(Equal(`{"id":1,"name":"Juuso","email":"foo@example.com","posts":[]}`))
			})

		})

		Context("creating second user with same email", func() {

			It("should return HTTP 422", func() {
				payload := `{"name": "Juuso", "password": "foo", "email": "foo@example.com"}`
				request, err := http.NewRequest("POST", "/api/user", strings.NewReader(payload))
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(422))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Email already in use"}`))
			})

		})

		Context("reading", func() {

			It("should shown up when requesting by ID", func() {
				request, err := http.NewRequest("GET", "/api/user/1", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				Expect(recorder.Body.String()).To(Equal(`{"id":1,"name":"Juuso","email":"foo@example.com","posts":[]}`))
			})

			It("non-existent ID should return not found", func() {
				request, err := http.NewRequest("GET", "/api/user/3", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(404))
				Expect(recorder.Body.String()).To(Equal(`{"error":"User not found"}`))
			})

			It("should be then listed on /users", func() {
				request, err := http.NewRequest("GET", "/api/users", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`[{"id":1,"name":"Juuso","email":"foo@example.com","posts":[]}]`))
			})
		})

		Context("accessing control panel before signing", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("GET", "/user", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
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
				Expect(recorder.Body.String()).To(Equal(`{"id":1,"name":"Juuso","email":"foo@example.com","posts":[]}`))
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
				payload := `{"title": "First post", "content": "This is example post with HTML elements like <b>bold</b> and <i>italics</i> in place."}`
				request, err := http.NewRequest("POST", "/api/post", strings.NewReader(payload))
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
				Expect(post.ID).To(Equal(int64(1)))
				Expect(post.Title).To(Equal("First post"))
				Expect(post.Content).To(Equal(`This is example post with HTML elements like <b>bold</b> and <i>italics</i> in place.`))
				Expect(post.Markdown).To(Equal(""))
				Expect(post.Slug).To(Equal("first-post"))
				Expect(post.Author).To(Equal(int64(1)))
				Expect(post.Date).Should(BeNumerically(">", int64(0)))
				Expect(post.Excerpt).To(Equal("This is example post with HTML elements like bold and italics in place."))
				Expect(post.Viewcount).To(Equal(uint(0)))
				*globalpost = post
				flag.Set("postslug", post.Slug)
			})
		})

		Context("reading", func() {

			It("non-existent slug should return not found", func() {
				request, err := http.NewRequest("GET", "/api/post/foo", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(404))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Post not found"}`))
			})

			It("post which exists should return 200 OK", func() {
				request, err := http.NewRequest("GET", "/api/post/"+*postslug, nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				var post Post
				if err := json.Unmarshal(recorder.Body.Bytes(), &post); err != nil {
					panic(err)
				}
				Expect(post).To(Equal(*globalpost))
				globalpost.Viewcount = uint(globalpost.Viewcount + 1)
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
				Expect(recorder.Body.String()).To(Equal(`{"success":"Post published"}`))
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
				Expect(sel).To(Equal("First post"))
			})

			It("after publishing, the post should be displayed trough API", func() {
				request, err := http.NewRequest("GET", "/api/posts", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				var posts []Post
				if err := json.Unmarshal(recorder.Body.Bytes(), &posts); err != nil {
					panic(err)
				}
				for i, post := range posts {
					Expect(i).To(Equal(0))
					Expect(post).To(Equal(*globalpost))
				}
			})
		})

		Context("owner of the first post", func() {

			It("should have the post listed in their account", func() {
				request, err := http.NewRequest("GET", "/api/user/1", nil)
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				var user User
				if err := json.Unmarshal(recorder.Body.Bytes(), &user); err != nil {
					panic(err)
				}
				Expect(user.ID).To(Equal(int64(1)))
				Expect(user.Name).To(Equal("Juuso"))
				Expect(user.Email).To(Equal("foo@example.com"))
				Expect(user.Posts[0]).To(Equal(*globalpost))
			})
		})

		Context("updating", func() {

			It("should return the updated post structure", func() {
				request, err := http.NewRequest("POST", "/api/post/"+*postslug+"/edit", strings.NewReader(`{"title": "First post edited", "content": "This is an EDITED example post with HTML elements like <b>bold</b> and <i>italics</i> in place."}`))
				if err != nil {
					panic(err)
				}
				globalpost.Title = "First post edited"
				globalpost.Content = "This is an EDITED example post with HTML elements like <b>bold</b> and <i>italics</i> in place."
				globalpost.Excerpt = "This is an EDITED example post with HTML elements like bold and italics in place."
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				var post Post
				if err := json.Unmarshal(recorder.Body.Bytes(), &post); err != nil {
					panic(err)
				}
				Expect(post).To(Equal(*globalpost))
			})

			It("after updating, the post should not be displayed on frontpage", func() {
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
				Expect(sel).To(Equal(""))
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
				var post Post
				if err := json.Unmarshal(recorder.Body.Bytes(), &post); err != nil {
					panic(err)
				}
				Expect(post).To(Equal(*globalpost))
			})
		})

		Context("creating second post", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("POST", "/api/post", strings.NewReader(`{"title": "Second post", "content": "This is second post"}`))
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
				Expect(post.ID).To(Equal(int64(2)))
				Expect(post.Title).To(Equal("Second post"))
				Expect(post.Content).To(Equal(`This is second post`))
				Expect(post.Markdown).To(Equal(""))
				Expect(post.Slug).To(Equal("second-post"))
				Expect(post.Author).To(Equal(int64(1)))
				Expect(post.Date).Should(BeNumerically(">", int64(0)))
				Expect(post.Excerpt).To(Equal("This is second post"))
				Expect(post.Viewcount).To(Equal(uint(0)))
				*globalpost = post
				flag.Set("postslug", post.Slug)
			})
		})

		Context("updating second post", func() {

			It("should return the updated post structure", func() {
				request, err := http.NewRequest("POST", "/api/post/"+*postslug+"/edit", strings.NewReader(`{"title": "Second post edited", "content": "This is edited second post"}`))
				if err != nil {
					panic(err)
				}
				globalpost.Title = "Second post edited"
				globalpost.Content = "This is edited second post"
				globalpost.Excerpt = "This is edited second post"
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				var post Post
				if err := json.Unmarshal(recorder.Body.Bytes(), &post); err != nil {
					panic(err)
				}
				Expect(post).To(Equal(*globalpost))
			})

		})

		Context("reading posts on user control panel", func() {

			It("should list both of them", func() {
				request, err := http.NewRequest("GET", "/user", nil)
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}
				doc.Find("ul").Each(func(i int, s *goquery.Selection) {
					if i == 0 {
						Expect(s.Find("li a").First().Text()).To(Equal("First post edited"))
					}
					if i == 1 {
						Expect(s.Find("li a").First().Text()).To(Equal("Second post edited"))
					}
				})
			})

		})

		Context("creating third post", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("POST", "/api/post", strings.NewReader(`{"title": "Third post", "content": "This is second post"}`))
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

		Context("publishing third post", func() {

			It("with session data should return HTTP 200", func() {
				request, err := http.NewRequest("GET", "/api/post/third-post/publish", nil)
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

		})

		Context("reading third post after publishing", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("GET", "/api/post/"+*postslug, nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

		})

		Context("reading all three posts on user control panel", func() {

			It("should list all three of them", func() {
				request, err := http.NewRequest("GET", "/user", nil)
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}
				doc.Find("ul").Each(func(i int, s *goquery.Selection) {
					Expect(i).Should(BeNumerically("<=", 2))
				})
			})
		})

		Context("deleting third post", func() {

			It("without sessioncookies it should return 401", func() {
				request, err := http.NewRequest("GET", "/api/post/"+*postslug+"/delete", nil)
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
			})

			It("with sessioncookies it should return 200", func() {
				request, err := http.NewRequest("GET", "/api/post/"+*postslug+"/delete", nil)
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("should after deletion, only list two posts on user control panel", func() {
				request, err := http.NewRequest("GET", "/user", nil)
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}
				doc.Find("ul").Each(func(i int, s *goquery.Selection) {
					if i == 0 {
						Expect(s.Find("li a").First().Text()).To(Equal("First post edited"))
					}
					if i == 1 {
						Expect(s.Find("li a").First().Text()).To(Equal("Second post edited"))
					}
				})
			})
		})

		Context("Settings on /user/settings", func() {

			It("reading without sessioncookies it should return 401", func() {
				request, err := http.NewRequest("GET", "/api/settings", nil)
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
			})

			It("reading with sessioncookies it should return 200", func() {
				request, err := http.NewRequest("GET", "/api/settings", nil)
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`{"name":"Juuso's Blog","hostname":"example.com","allowregistrations":true,"markdown":false,"description":"Foo's test blog","mailgun":{"mgdomain":"foo","mgprikey":"foo"}}`))
				Expect(recorder.Code).To(Equal(200))
			})

			It("updaring without sessioncookie", func() {
				request, err := http.NewRequest("POST", "/api/settings", strings.NewReader(`{"name":"Juuso's Blog","hostname":"example.com","allowregistrations":false,"description":"Foo's test blog","mailgun":{"mgdomain":"foo","mgprikey":"foo"}}`))
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
			})

			It("updaring without sessioncookie", func() {
				request, err := http.NewRequest("POST", "/api/settings", strings.NewReader(`{"name":"Juuso's Blog","hostname":"example.com","allowregistrations":false,"description":"Foo's test blog","mailgun":{"mgdomain":"foo","mgprikey":"foo"}}`))
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`{"success":"Settings were successfully saved"}`))
				Expect(recorder.Code).To(Equal(200))
			})

			It("reading with sessioncookies it should return 200", func() {
				request, err := http.NewRequest("GET", "/api/settings", nil)
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`{"name":"Juuso's Blog","hostname":"example.com","allowregistrations":false,"markdown":false,"description":"Foo's test blog","mailgun":{"mgdomain":"foo","mgprikey":"foo"}}`))
				Expect(recorder.Code).To(Equal(200))
			})
		})

	})

	Describe("Users", func() {

		Context("creation", func() {

			It("should return HTTP 403 because allowregistrations is false", func() {
				request, err := http.NewRequest("POST", "/api/user", strings.NewReader(`{"name": "Juuso", "password": "hello", "email": "bar@example.com"}`))
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(403))
			})

		})

	})

	Describe("Markdown", func() {

		Context("switching to Markdown", func() {

			It("changing settings should return HTTP 200", func() {
				request, err := http.NewRequest("POST", "/api/settings", strings.NewReader(`{"name":"Juuso's Blog","hostname":"example.com","allowregistrations":false,"markdown":true,"description":"Foo's test blog","mailgun":{"mgdomain":"foo","mgprikey":"foo"}}`))
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("should change global Settings variable", func() {
				request, err := http.NewRequest("GET", "/api/settings", nil)
				if err != nil {
					panic(err)
				}
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`{"name":"Juuso's Blog","hostname":"example.com","allowregistrations":false,"markdown":true,"description":"Foo's test blog","mailgun":{"mgdomain":"foo","mgprikey":"foo"}}`))
				Expect(recorder.Code).To(Equal(200))
				Expect(Settings.Markdown).To(Equal(true))
				Expect(Settings.AllowRegistrations).To(Equal(false))
			})
		})

		Context("posts", func() {

			It("creating one should return 200", func() {
				request, err := http.NewRequest("POST", "/api/post", strings.NewReader(`{"title": "Markdown post", "markdown": "### foo\n*foo* foo **foo**"}`))
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
				Expect(post.ID).To(Equal(int64(3)))
				Expect(post.Title).To(Equal("Markdown post"))
				Expect(post.Content).To(Equal("\u003ch3\u003efoo\u003cem\u003efoo\u003c/em\u003e foo \u003cstrong\u003efoo\u003c/strong\u003e\u003c/h3\u003e\n"))
				Expect(post.Markdown).To(Equal("### foo\n*foo* foo **foo**"))
				Expect(post.Slug).To(Equal("markdown-post"))
				Expect(post.Author).To(Equal(int64(1)))
				Expect(post.Date).Should(BeNumerically(">", int64(0)))
				Expect(post.Viewcount).To(Equal(uint(0)))
				*globalpost = post
				flag.Set("postslug", post.Slug)
			})

		})

	})

})
