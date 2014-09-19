package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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

		})
	})
})
