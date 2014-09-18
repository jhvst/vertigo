package main

import (
	"net/http"
	"net/http/httptest"
	"os"

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
		})
	})

	Describe("Settings", func() {

		Context("after creation", func() {

			It("Firstrun should equal to true", func() {
				settings := VertigoSettings()
				Expect(settings.Firstrun).To(Equal(true))
			})

		})
	})
})
