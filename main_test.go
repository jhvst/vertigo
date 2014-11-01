package vertigo

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
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

		Context("loading the homepage", func() {
			It("should display installation wizard", func() {
				request, _ := http.NewRequest("GET", "/", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}
				sel := doc.Find("h1").First().Text()
				Expect(sel).To(Equal("Your settings file seems to miss some fields. Lets fix that."))
			})
		})
	})

	Describe("Static pages", func() {
		AfterEach(func() {
			server.ServeHTTP(recorder, request)
			Expect(recorder.Code).To(Equal(200))
		})

		Context("API index page", func() {
			It("should respond 200 OK", func() {
				request, _ = http.NewRequest("GET", "/api", nil)
			})
		})

		Context("User register", func() {
			It("should respond 200 OK", func() {
				request, _ = http.NewRequest("GET", "/user/register", nil)
			})
		})

		Context("User recover", func() {
			It("should respond 200 OK", func() {
				request, _ = http.NewRequest("GET", "/user/recover", nil)
			})
		})

		Context("User recover", func() {
			It("should respond 200 OK", func() {
				request, _ = http.NewRequest("GET", "/user/login", nil)
			})
		})

		Context("User reset", func() {
			It("should respond 200 OK", func() {
				request, _ = http.NewRequest("GET", "/user/reset/1/aaa", nil)
			})
		})
	})

	Describe("Empty public API routes", func() {

		Context("pages with arrays", func() {

			AfterEach(func() {
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				Expect(recorder.Body.String()).To(Equal("[]"))
			})

			It("should respond with []", func() {
				request, _ = http.NewRequest("GET", "/api/posts", nil)
			})

			It("should respond with []", func() {
				request, _ = http.NewRequest("GET", "/api/users", nil)
			})
		})

		Context("pages with single item", func() {

			AfterEach(func() {
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(404))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Not found"}`))
			})

			It("should respond with []", func() {
				request, _ = http.NewRequest("GET", "/api/post/0", nil)
			})

			It("should respond with []", func() {
				request, _ = http.NewRequest("GET", "/api/user/0", nil)
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

			var settings Vertigo

			It("response should be a redirection", func() {
				settings.Hostname = "example.com"
				settings.Name = "Foo's blog"
				settings.Description = "Foo's test blog"
				settings.Mailer.Domain = os.Getenv("MAILGUN_API_DOMAIN")
				settings.Mailer.PrivateKey = os.Getenv("MAILGUN_API_KEY")
				payload, _ := json.Marshal(settings)
				request, _ := http.NewRequest("POST", "/api/installation", bytes.NewReader(payload))
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("the settings.json should have all fields populated", func() {
				Expect(Settings.Hostname).To(Equal(settings.Hostname))
				Expect(Settings.Name).To(Equal(settings.Name))
				Expect(Settings.Description).To(Equal(settings.Description))
				Expect(Settings.Mailer.Domain).To(Equal(settings.Mailer.Domain))
				Expect(Settings.Mailer.PrivateKey).To(Equal(settings.Mailer.PrivateKey))
				Expect(Settings.AllowRegistrations).To(Equal(true))
				Expect(Settings.Markdown).To(Equal(false))
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
				Expect(Settings.Name).To(Equal(settings.Name))
			})

			It("frontpage's <title> should now be 'Juuso's Blog'", func() {
				request, _ := http.NewRequest("GET", "/", nil)
				server.ServeHTTP(recorder, request)
				doc, _ := goquery.NewDocumentFromReader(recorder.Body)
				sel := doc.Find("title").Text()
				Expect(sel).To(Equal(Settings.Name))
			})
		})
	})

	Describe("Users", func() {

		Context("creation", func() {

			It("should return HTTP 200", func() {
				payload := `{"name": "Juuso", "password": "foo", "email": "vertigo-test@mailinator.com"}`
				request, _ := http.NewRequest("POST", "/api/user", strings.NewReader(payload))
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				Expect(recorder.Body.String()).To(Equal(`{"id":1,"name":"Juuso","email":"vertigo-test@mailinator.com","posts":[]}`))
			})
		})

		Context("creating second user with same email", func() {

			It("should return HTTP 422", func() {
				payload := `{"name": "Juuso", "password": "foo", "email": "vertigo-test@mailinator.com"}`
				request, _ := http.NewRequest("POST", "/api/user", strings.NewReader(payload))
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(422))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Email already in use"}`))
			})
		})

		Context("reading", func() {

			It("should shown up when requesting by ID", func() {
				request, _ := http.NewRequest("GET", "/api/user/1", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				Expect(recorder.Body.String()).To(Equal(`{"id":1,"name":"Juuso","email":"vertigo-test@mailinator.com","posts":[]}`))
			})

			It("non-existent ID should return not found", func() {
				request, _ := http.NewRequest("GET", "/api/user/3", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(404))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Not found"}`))
			})

			It("should be then listed on /users", func() {
				request, _ := http.NewRequest("GET", "/api/users", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`[{"id":1,"name":"Juuso","email":"vertigo-test@mailinator.com","posts":[]}]`))
			})
		})

		Context("accessing control panel before signing", func() {

			It("should return HTTP 200", func() {
				request, _ := http.NewRequest("GET", "/user", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
			})
		})

		Context("signing in", func() {

			It("should return HTTP 200", func() {
				request, _ := http.NewRequest("POST", "/api/user/login", strings.NewReader(`{"name": "Juuso", "password": "foo", "email": "vertigo-test@mailinator.com"}`))
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				// i assure, nothing else worked
				cookie := strings.Split(strings.TrimLeft(recorder.HeaderMap["Set-Cookie"][0], "user="), ";")[0]
				flag.Set("sessioncookie", cookie)
				fmt.Println("User sessioncookie:", *sessioncookie)
				Expect(recorder.Code).To(Equal(200))
				Expect(recorder.Body.String()).To(Equal(`{"id":1,"name":"Juuso","email":"vertigo-test@mailinator.com","posts":[]}`))
			})
		})

		Context("accessing control panel after signing", func() {

			It("should return HTTP 200", func() {
				request, _ := http.NewRequest("GET", "/user", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})
		})

		Context("accessing sessionredirected page like registration after signing", func() {

			It("should return HTTP 302", func() {
				request, _ := http.NewRequest("GET", "/user/login", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(302))
			})
		})
	})

	Describe("Posts", func() {

		Context("loading the creation page", func() {
			It("should return HTTP 200", func() {
				request, _ := http.NewRequest("GET", "/post/new", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})
		})

		Context("creation", func() {

			It("should return HTTP 200", func() {
				payload := `{"title": "First post", "content": "This is example post with HTML elements like <b>bold</b> and <i>italics</i> in place."}`
				request, _ := http.NewRequest("POST", "/api/post", strings.NewReader(payload))
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
				request, _ := http.NewRequest("GET", "/api/post/foo", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(404))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Not found"}`))
			})

			It("post which exists should return 200 OK", func() {
				request, _ := http.NewRequest("GET", "/api/post/"+*postslug, nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				var post Post
				if err := json.Unmarshal(recorder.Body.Bytes(), &post); err != nil {
					panic(err)
				}
				Expect(post).To(Equal(*globalpost))
				globalpost.Viewcount = uint(globalpost.Viewcount + 1)
			})

			It("on frontend, post which exists should return 200 OK", func() {
				request, _ := http.NewRequest("GET", "/post/"+*postslug, nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				globalpost.Viewcount = uint(globalpost.Viewcount + 1)
			})
		})

		Context("publishing", func() {

			It("without session data should return HTTP 401", func() {
				request, _ := http.NewRequest("GET", "/api/post/"+*postslug+"/publish", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
			})

			It("with session data should return HTTP 200", func() {
				request, _ := http.NewRequest("GET", "/api/post/"+*postslug+"/publish", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`{"success":"Post published"}`))
				Expect(recorder.Code).To(Equal(200))
			})

			It("after publishing, the post should be displayed on frontpage", func() {
				request, _ := http.NewRequest("GET", "/", nil)
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
				request, _ := http.NewRequest("GET", "/api/posts", nil)
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
				request, _ := http.NewRequest("GET", "/api/user/1", nil)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				var user User
				if err := json.Unmarshal(recorder.Body.Bytes(), &user); err != nil {
					panic(err)
				}
				Expect(user.ID).To(Equal(int64(1)))
				Expect(user.Name).To(Equal("Juuso"))
				Expect(user.Email).To(Equal("vertigo-test@mailinator.com"))
				Expect(user.Posts[0]).To(Equal(*globalpost))
			})
		})

		Context("updating", func() {

			It("loading frontend edit page should respond 200 OK", func() {
				request, _ := http.NewRequest("GET", "/post/"+*postslug+"/edit", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("should return 401 without authorization", func() {
				request, _ := http.NewRequest("POST", "/api/post/"+*postslug+"/edit", strings.NewReader(`{"title": "First post edited", "content": "This is an EDITED example post with HTML elements like <b>bold</b> and <i>italics</i> in place."}`))
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Unauthorized"}`))
			})

			It("should return 401 without malformed authorization", func() {
				request, _ := http.NewRequest("POST", "/api/post/"+*postslug+"/edit", strings.NewReader(`{"title": "First post edited", "content": "This is an EDITED example post with HTML elements like <b>bold</b> and <i>italics</i> in place."}`))
				cookie := &http.Cookie{Name: "user", Value: "MTQxNDc2NzAyOXxEdi1CQkFFQ180SUFBUkFCRUFBQUhmLUNBQUVHYzNSeWFXNW5EQVlBQkhWelpYSUZhVzUwTmpRRUFnQUN8Y2PFc-lZ8aEMWypbKXTD-LWg6o9DtJaMzd8NMc8m87A="}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Unauthorized"}`))
			})

			It("should return 404 with non-existent post", func() {
				request, _ := http.NewRequest("POST", "/api/post/foobar/edit", strings.NewReader(`{"title": "First post edited", "content": "This is an EDITED example post with HTML elements like <b>bold</b> and <i>italics</i> in place."}`))
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(404))
				Expect(recorder.Body.String()).To(Equal(`{"error":"Not found"}`))
			})

			It("should return the updated post structure", func() {
				request, _ := http.NewRequest("POST", "/api/post/"+*postslug+"/edit", strings.NewReader(`{"title": "First post edited", "content": "This is an EDITED example post with HTML elements like <b>bold</b> and <i>italics</i> in place."}`))
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
				request, _ := http.NewRequest("GET", "/", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				doc, _ := goquery.NewDocumentFromReader(recorder.Body)
				sel := doc.Find("article h1").Text()
				Expect(sel).To(Equal(""))
			})

			It("after updating, the post should not be displayed trough API", func() {
				request, _ := http.NewRequest("GET", "/api/posts", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				Expect(recorder.Body.String()).To(Equal("[]"))
			})
		})

		Context("reading after updating", func() {

			It("should return HTTP 200", func() {
				request, _ := http.NewRequest("GET", "/api/post/"+*postslug, nil)
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
				request, _ := http.NewRequest("POST", "/api/post", strings.NewReader(`{"title": "Second post", "content": "This is second post"}`))
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
				request, _ := http.NewRequest("POST", "/api/post/"+*postslug+"/edit", strings.NewReader(`{"title": "Second post edited", "content": "This is edited second post"}`))
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
				request, _ := http.NewRequest("GET", "/user", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}
				doc.Find("ul").Each(func(i int, s *goquery.Selection) {
					Expect(i).Should(BeNumerically("<=", 1))
				})
			})
		})

		Context("creating third post", func() {

			It("should return HTTP 200", func() {
				request, _ := http.NewRequest("POST", "/api/post", strings.NewReader(`{"title": "Third post", "content": "This is second post"}`))
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
				request, _ := http.NewRequest("GET", "/api/post/third-post/publish", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})
		})

		Context("reading third post after publishing", func() {

			It("should return HTTP 200", func() {
				request, _ := http.NewRequest("GET", "/api/post/"+*postslug, nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})
		})

		Context("reading all three posts on user control panel", func() {

			It("should list all three of them", func() {
				request, _ := http.NewRequest("GET", "/user", nil)
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
				request, _ := http.NewRequest("GET", "/api/post/"+*postslug+"/delete", nil)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
			})

			It("with sessioncookies it should return 200", func() {
				request, _ := http.NewRequest("GET", "/api/post/"+*postslug+"/delete", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("should after deletion, only list two posts on user control panel", func() {
				request, _ := http.NewRequest("GET", "/user", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}
				doc.Find("ul").Each(func(i int, s *goquery.Selection) {
					Expect(i).Should(BeNumerically("<=", 1))
				})
			})
		})

		Context("Settings on /user/settings", func() {

			It("reading without sessioncookies it should return 401", func() {
				request, _ := http.NewRequest("GET", "/api/settings", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
			})

			It("reading with sessioncookies it should return 200", func() {
				request, _ := http.NewRequest("GET", "/api/settings", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				var settings Vertigo
				if err := json.Unmarshal(recorder.Body.Bytes(), &settings); err != nil {
					panic(err)
				}
				var settingsWithoutCookieHash Vertigo
				settingsWithoutCookieHash = *Settings
				settingsWithoutCookieHash.CookieHash = ""
				Expect(settings).To(Equal(settingsWithoutCookieHash))
				Expect(recorder.Code).To(Equal(200))
			})

			It("frontend, reading with sessioncookies it should return 200", func() {
				request, _ := http.NewRequest("GET", "/user/settings", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("updating without sessioncookie", func() {
				request, _ := http.NewRequest("POST", "/api/settings", strings.NewReader(`{"name":"Juuso's Blog","hostname":"example.com","allowregistrations":false,"markdown":false,"description":"Foo's test blog","mailgun":{"mgdomain":"foo","mgprikey":"foo"}}`))
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
			})

			It("updating with sessioncookie", func() {
				var settings Vertigo
				settings = *Settings
				settings.Name = "Foo's Blog"
				settings.AllowRegistrations = false
				payload, _ := json.Marshal(settings)
				request, _ := http.NewRequest("POST", "/api/settings", bytes.NewReader(payload))
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`{"success":"Settings were successfully saved"}`))
				Expect(recorder.Code).To(Equal(200))
			})

			It("reading with sessioncookies it should return 200", func() {
				request, _ := http.NewRequest("GET", "/api/settings", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				var settings Vertigo
				if err := json.Unmarshal(recorder.Body.Bytes(), &settings); err != nil {
					panic(err)
				}
				var settingsWithoutCookieHash Vertigo
				settingsWithoutCookieHash = *Settings
				settingsWithoutCookieHash.CookieHash = ""
				Expect(settings).To(Equal(settingsWithoutCookieHash))
				Expect(recorder.Code).To(Equal(200))
			})
		})
	})

	Describe("Users", func() {

		Context("creation", func() {

			It("should return HTTP 403 because allowregistrations is false", func() {
				request, _ := http.NewRequest("POST", "/api/user", strings.NewReader(`{"name": "Juuso", "password": "hello", "email": "bar@example.com"}`))
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(403))
			})
		})
	})

	Describe("Markdown", func() {

		Context("switching to Markdown", func() {

			It("changing settings should return HTTP 200", func() {
				var settings Vertigo
				settings = *Settings
				settings.Markdown = true
				payload, _ := json.Marshal(settings)
				request, _ = http.NewRequest("POST", "/api/settings", bytes.NewReader(payload))
				request.Header.Set("Content-Type", "application/json")
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("should change global Settings variable", func() {
				request, _ = http.NewRequest("GET", "/api/settings", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(Settings.Markdown).To(Equal(true))
				Expect(Settings.AllowRegistrations).To(Equal(false))
				Expect(recorder.Code).To(Equal(200))
			})
		})

		Context("posts", func() {

			It("creating one should return 200", func() {
				request, _ := http.NewRequest("POST", "/api/post", strings.NewReader(`{"title": "Markdown post", "markdown": "### foo\n*foo* foo **foo**"}`))
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
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
				Expect(recorder.Code).To(Equal(200))
			})
		})

		Context("publishing", func() {

			It("should return error on non-existent ID", func() {
				request, _ := http.NewRequest("GET", "/api/post/foo-post/publish", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`{"error":"Not found"}`))
				Expect(recorder.Code).To(Equal(404))
			})

			It("should return error without authentication", func() {
				request, _ := http.NewRequest("GET", "/api/post/markdown-post/publish", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`{"error":"Unauthorized"}`))
				Expect(recorder.Code).To(Equal(401))
			})

			It("with session data should return HTTP 200", func() {
				request, _ := http.NewRequest("GET", "/api/post/markdown-post/publish", nil)
				cookie := &http.Cookie{Name: "user", Value: *sessioncookie}
				request.AddCookie(cookie)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal(`{"success":"Post published"}`))
				Expect(recorder.Code).To(Equal(200))
			})
		})
	})

	Describe("Feeds", func() {

		Context("reading feeds without defining feed type", func() {

			It("should redirect to RSS", func() {
				request, _ := http.NewRequest("GET", "/feeds", nil)
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(302))
				Expect(recorder.HeaderMap["Location"][0]).To(Equal("/feeds/rss"))
			})
		})

		Context("when defining feed", func() {

			AfterEach(func() {
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
				Expect(recorder.HeaderMap["Content-Type"][0]).To(Equal("application/xml"))
			})

			It("should be RSS in /rss", func() {
				request, _ = http.NewRequest("GET", "/feeds/rss", nil)
			})

			It("should be Atom in /atom", func() {
				request, _ = http.NewRequest("GET", "/feeds/atom", nil)
			})
		})
	})

	Describe("Search", func() {

		Context("searching for the published Markdown post", func() {

			AfterEach(func() {
				Expect(recorder.Code).To(Equal(200))
			})

			It("should return it", func() {
				request, _ := http.NewRequest("POST", "/api/post/search", strings.NewReader(`{"query": "markdown"}`))
				request.Header.Set("Content-Type", "application/json")
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

		Context("searching with a query which is not contained in any post", func() {

			It("should return empty array in JSON", func() {
				request, _ := http.NewRequest("POST", "/api/post/search", strings.NewReader(`{"query": "foofoobarbar"}`))
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Body.String()).To(Equal("[]"))
			})
		})

		Context("searching for the published Markdown post on frontend", func() {

			It("should return it", func() {
				request, _ := http.NewRequest("POST", "/post/search", strings.NewReader(`query=markdown`))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				server.ServeHTTP(recorder, request)
				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}
				sel := doc.Find("h1").Text()
				Expect(sel).To(Equal("Markdown post"))
			})
		})

		Context("searching with a query which is not contained in any post on frontend", func() {

			It("should display nothing found page", func() {
				request, _ := http.NewRequest("POST", "/post/search", strings.NewReader(`query=foofoobarbar`))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				server.ServeHTTP(recorder, request)
				doc, err := goquery.NewDocumentFromReader(recorder.Body)
				if err != nil {
					panic(err)
				}
				sel := doc.Find("h2").Text()
				Expect(sel).To(Equal("Nothing found."))
			})
		})
	})

	Describe("Users", func() {

		Context("logging out", func() {

			It("should return HTTP 200", func() {
				request, _ := http.NewRequest("GET", "/api/user/logout", nil)
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})
		})
	})

	Describe("Password recovery", func() {

		var recovery string

		Context("with email which does not exist", func() {

			It("should respond 401", func() {
				request, _ := http.NewRequest("POST", "/user/recover", strings.NewReader(`email=foobar@example.com`))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(401))
				Expect(recorder.Body.String()).To(Equal(`{"error":"User with that email does not exist."}`))
			})
		})

		Context("with email which does exist", func() {

			It("should redirect to login page with notification", func() {
				request, _ := http.NewRequest("POST", "/user/recover", strings.NewReader(`email=vertigo-test@mailinator.com`))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(302))
			})
		})

		Context("the user structure", func() {

			It("should have the recovery key in place", func() {
				db, err := gorm.Open(os.Getenv("driver"), os.Getenv("dbsource"))
				if err != nil {
					panic(err)
				}

				var user User
				user.Email = "vertigo-test@mailinator.com"
				user, err = user.GetByEmail(&db)
				if err != nil {
					panic(err)
				}
				recovery = user.Recovery
				Expect(user.Recovery).ShouldNot(Equal(""))
			})
		})

		Context("resetting the password", func() {

			It("the route should redirect", func() {
				request, _ := http.NewRequest("POST", "/user/reset/1/"+recovery, strings.NewReader(`password=newpassword`))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(302))
			})

			It("should not have the recovery key in place", func() {
				db, err := gorm.Open(os.Getenv("driver"), os.Getenv("dbsource"))
				if err != nil {
					panic(err)
				}

				var user User
				user.Email = "vertigo-test@mailinator.com"
				user, err = user.GetByEmail(&db)
				if err != nil {
					panic(err)
				}
				recovery = user.Recovery
				Expect(user.Recovery).To(Equal(" "))
			})

			It("the login should now work", func() {
				request, _ := http.NewRequest("POST", "/user/login", strings.NewReader(`email=vertigo-test@mailinator.com&password=newpassword`))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				server.ServeHTTP(recorder, request)
				Expect(recorder.HeaderMap["Location"][0]).To(Equal("/user"))
				Expect(recorder.Code).To(Equal(302))
			})
		})

		Context("recovery key expiration", func() {

			It("should redirect to login page with notification", func() {
				request, _ := http.NewRequest("POST", "/user/recover", strings.NewReader(`email=vertigo-test@mailinator.com`))
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(302))
			})
		})

		Context("the user structure", func() {

			It("should have the recovery key in place", func() {
				db, err := gorm.Open(os.Getenv("driver"), os.Getenv("dbsource"))
				if err != nil {
					panic(err)
				}

				var user User
				user.Email = "vertigo-test@mailinator.com"
				user, err = user.GetByEmail(&db)
				if err != nil {
					panic(err)
				}
				recovery = user.Recovery
				Expect(user.Recovery).ShouldNot(Equal(" "))
			})

			It("the recovery key should expire in 1 second", func() {

				db, err := gorm.Open(os.Getenv("driver"), os.Getenv("dbsource"))
				if err != nil {
					panic(err)
				}

				var user User
				user.ID = int64(1)
				go user.ExpireRecovery(&db, 1*time.Second)
				time.Sleep(2 * time.Second)

				db, err = gorm.Open(os.Getenv("driver"), os.Getenv("dbsource"))
				if err != nil {
					panic(err)
				}

				user, err = user.Get(&db)
				if err != nil {
					panic(err)
				}
				recovery = user.Recovery
				Expect(user.Recovery).Should(Equal(" "))
			})
		})
	})

})
