package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var sessioncookie *string = flag.String("sessioncookie", "", "global flag for test sessioncookie")

var _ = Describe("Vertigo", func() {

	var server Server
	var request *http.Request
	var recorder *httptest.ResponseRecorder

	if os.Getenv("WERCKER_RETHINKDB_HOST") != "" {
		os.Setenv("RDB_HOST", os.Getenv("WERCKER_RETHINKDB_HOST"))
	}
	if os.Getenv("WERCKER_RETHINKDB_PORT") != "" {
		os.Setenv("RDB_PORT", os.Getenv("WERCKER_RETHINKDB_PORT"))
	}

	BeforeEach(func() {
		// Set up a new server, connected to a test database,
		// before each test.
		server = NewServer()

		// Record HTTP responses.
		recorder = httptest.NewRecorder()
	})

	Describe("GET / (homepage)", func() {

		// Set up a new GET request before every test
		// in this describe block.
		BeforeEach(func() {
			request, _ = http.NewRequest("GET", "/", nil)
		})

		Context("", func() {
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

	Describe("Creating a user", func() {

		Context("POSTing to /api/user", func() {

			It("should return HTTP 200", func() {
				request, err := http.NewRequest("POST", "/api/user", strings.NewReader(`{"name": "Juuso", "password": "foo", "email": "foo@example.com"}`))
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				Expect(recorder.Code).To(Equal(200))
			})

			It("should be then listed on /users", func() {
				request, err := http.NewRequest("GET", "/api/users", nil)
				if err != nil {
					panic(err)
				}
				server.ServeHTTP(recorder, request)
				var users []Person
				if err := json.Unmarshal(recorder.Body.Bytes(), &users); err != nil {
					panic(err)
				}
				fmt.Println("User structs listed on /users", recorder.Body)
				Expect(recorder.Code).To(Equal(200))
				for i, user := range users {
					Expect(i).To(Equal(0))
					Expect(user.Name).To(Equal("Juuso"))
					Expect(user.ID).NotTo(Equal(""))
				}
			})
		})
	})

	Describe("Logging in a user", func() {

		Context("POSTing to /api/user/login", func() {

			It("should return HTTP 200", func() {

				request, err := http.NewRequest("POST", "/api/user/login", strings.NewReader(`{"name": "Juuso", "password": "foo", "email": "foo@example.com"}`))
				if err != nil {
					panic(err)
				}
				request.Header.Set("Content-Type", "application/json")
				server.ServeHTTP(recorder, request)
				flag.Set("sessioncookie", strings.Split(strings.Split(recorder.HeaderMap["Set-Cookie"][0], ";")[0], "=")[1])
				fmt.Println("User sessioncookie:", *sessioncookie)
				fmt.Println("User struct responded in login", recorder.Body)
				Expect(recorder.Code).To(Equal(200))

			})
		})

		Context("Accessing control panel", func() {

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

})
