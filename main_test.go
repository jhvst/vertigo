package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/toldjuuso/vertigo/databases/sqlx"

	"github.com/PuerkitoBio/goquery"
	"github.com/russross/blackfriday"
	slug "github.com/shurcooL/sanitized_anchor_name"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/toldjuuso/excerpt"
)

var server = NewServer()
var settings Vertigo
var user User
var post Post
var userid string
var postslug string
var recovery string
var sessioncookie string
var secondusersessioncookie string
var malformedsessioncookie = "MTQxNDc2NzAyOXxEdi1CQkFFQ180SUFBUkFCRUFBQUhmLUNBQUVHYzNSeWFXNW5EQVlBQkhWelpYSUZhVzUwTmpRRUFnQUN8Y2PFc-lZ8aEMWypbKXTD-LWg6o9DtJaMzd8NMc8m87A="

func TestInstallationWizard(t *testing.T) {

	Convey("Opening homepage", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/", nil)

		Convey("it should display installation wizard", func() {
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			doc, _ := goquery.NewDocumentFromReader(recorder.Body)
			sel := doc.Find("h1").First().Text()
			So(sel, ShouldEqual, "Your settings file seems to be missing some fields. Lets fix that.")
		})
	})
}

func TestStaticPages(t *testing.T) {

	Convey("All static pages should return 200 OK", t, func() {
		var recorder = httptest.NewRecorder()

		Convey("Loading API index page", func() {
			request, _ := http.NewRequest("GET", "/api", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
		})

		Convey("Loading user register page", func() {
			request, _ := http.NewRequest("GET", "/user/register", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
		})

		Convey("Loading user login page", func() {
			request, _ := http.NewRequest("GET", "/user/login", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
		})

		Convey("Loading user recovery page", func() {
			request, _ := http.NewRequest("GET", "/user/recover", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
		})

		Convey("Loading user reset page", func() {
			request, _ := http.NewRequest("GET", "/user/reset/1/foobar", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
		})

		Convey("Loading non-existent page", func() {
			request, _ := http.NewRequest("GET", "/foobar", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 404)
		})
	})
}

func TestEmptyAPIRoutes(t *testing.T) {

	Convey("API routes should return empty JSON responses", t, func() {
		var recorder = httptest.NewRecorder()

		Convey("posts page should return []", func() {
			request, _ := http.NewRequest("GET", "/api/posts", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			So(recorder.Body.String(), ShouldEqual, "[]")
		})

		Convey("users page should return []", func() {
			request, _ := http.NewRequest("GET", "/api/users", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			So(recorder.Body.String(), ShouldEqual, "[]")
		})

		Convey("get post should return 404", func() {
			request, _ := http.NewRequest("GET", "/api/user/0", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 404)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Not found"}`)
		})

		Convey("get user should return 404", func() {
			request, _ := http.NewRequest("GET", "/api/post/0", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 404)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Not found"}`)
		})
	})
}

func TestCreatingSettings(t *testing.T) {

	Convey("after creating", t, func() {

		Convey("settings.Firstrun should equal true", func() {
			settings := VertigoSettings()
			So(settings.Firstrun, ShouldBeTrue)
		})
	})
}

func TestSavingSettingsViaInstallationWizard(t *testing.T) {

	Convey("by submitting data in JSON", t, func() {
		var recorder = httptest.NewRecorder()

		Convey("it should return 200 OK", func() {
			settings.Hostname = "http://example.com"
			settings.Name = "Foo's blog"
			settings.Description = "Foo's test blog"
			settings.MailerLogin = os.Getenv("SMTP_LOGIN")
			settings.MailerPassword = os.Getenv("SMTP_PASSWORD")
			port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
			settings.MailerPort = port
			settings.MailerHostname = os.Getenv("SMTP_SERVER")
			payload, _ := json.Marshal(settings)
			request, _ := http.NewRequest("POST", "/api/installation", bytes.NewReader(payload))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
		})
	})
}

func TestSettingValues(t *testing.T) {

	Convey("Settings object should be the same as settings object", t, func() {
		So(Settings.Hostname, ShouldEqual, settings.Hostname)
		So(Settings.Name, ShouldEqual, settings.Name)
		So(Settings.Description, ShouldEqual, settings.Description)
		So(Settings.MailerLogin, ShouldEqual, settings.MailerLogin)
		So(Settings.MailerPassword, ShouldEqual, settings.MailerPassword)
		So(Settings.MailerPort, ShouldEqual, settings.MailerPort)
		So(Settings.MailerHostname, ShouldEqual, settings.MailerHostname)
		So(Settings.AllowRegistrations, ShouldBeTrue)
	})
}

func TestManipulatingSettings(t *testing.T) {

	Convey("when manipulating the global Settings variable", t, func() {

		Convey("should save the changes to disk", func() {
			settings = *Settings
			settings.Name = "Juuso's Blog"
			s, err := settings.Update()
			if err != nil {
				panic(err)
			}
			Settings = s
		})

		Convey("frontpage's <title> should now be 'Juuso's Blog'", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/", nil)
			server.ServeHTTP(recorder, request)
			doc, _ := goquery.NewDocumentFromReader(recorder.Body)
			sel := doc.Find("title").Text()
			So(sel, ShouldEqual, Settings.Name)
		})
	})

	TestSettingValues(t)
}

func TestCreateFirstUser(t *testing.T) {

	user.Name = "Juuso"
	user.Password = "foo"
	user.Email = "vertigo-test@mailinator.com"
	user.Location = "Europe/Helsinki"
	testCreateUser(t, user.Name, user.Password, user.Email, user.Location)
}

func testCreateUser(t *testing.T, name string, password string, email string, location string) {

	// Convey("using frontend", t, func() {

	// 	Convey("with valid input it should return 200 OK", func() {
	// 		var recorder = httptest.NewRecorder()
	// 		payload := fmt.Sprintf(`name=%s&password=%s&email=%s`, name, password, email)
	// 		request, _ := http.NewRequest("POST", "/user/register", strings.NewReader(payload))
	// 		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 		server.ServeHTTP(recorder, request)
	// 		So(recorder.Code, ShouldEqual, 302)
	// 	})
	// })

	Convey("using API", t, func() {

		payload := fmt.Sprintf(`{"name":"%s", "password":"%s", "email":"%s", "location":"%s"}`, name, password, email, location)
		badpayload := fmt.Sprintf(`{"name":"%s", "password":"%s", "email":"%s", "location":"EU/FI"}`, name, password, email)

		Convey("with bad location it should should return 422", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/api/user", strings.NewReader(badpayload))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 422)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Location invalid. Please use IANA timezone database compatible locations."}`)
		})

		Convey("with valid input it should return 200 OK", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/api/user", strings.NewReader(payload))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			json.Unmarshal(recorder.Body.Bytes(), &user)
			So(user.Name, ShouldEqual, name)
			So(user.Email, ShouldEqual, email)
			So(user.Posts, ShouldBeEmpty)
			So(user.Location, ShouldEqual, location)
		})

		Convey("with the same email should return 422", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/api/user", strings.NewReader(payload))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 422)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Email already in use"}`)
		})
	})
}

func TestReadUser(t *testing.T) {

	Convey("reading the latest user by ID should return 200 OK", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", fmt.Sprintf("/api/user/%d", user.ID), nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		So(recorder.Body.String(), ShouldEqual, `{"id":1,"name":"Juuso","email":"vertigo-test@mailinator.com","posts":[],"location":"Europe/Helsinki"}`)
	})
}

func TestReadUsers(t *testing.T) {

	Convey("reading all users return 200 OK", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/api/users/", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		So(recorder.Body.String(), ShouldEqual, `[{"id":1,"name":"Juuso","email":"vertigo-test@mailinator.com","posts":[],"location":"Europe/Helsinki"}]`)
	})
}

func TestReadUserSpecialCases(t *testing.T) {

	Convey("when user does not exist, it should return not found", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/api/user/3", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 404)
		So(recorder.Body.String(), ShouldEqual, `{"error":"Not found"}`)
	})

	Convey("when user ID is malformed (eg. string), it should return 400", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/api/user/foobar", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 400)
		So(recorder.Body.String(), ShouldEqual, `{"error":"The user ID could not be parsed from the request URL."}`)
	})
}

func TestUserSignin(t *testing.T) {

	Convey("on frontend", t, func() {
		var recorder = httptest.NewRecorder()

		Convey("should return 401 with wrong password", func() {
			request, _ := http.NewRequest("POST", "/user/login", strings.NewReader(`password=foobar&email=vertigo-test@mailinator.com`))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
		})

		Convey("should return 404 with non-existent email", func() {
			request, _ := http.NewRequest("POST", "/user/login", strings.NewReader(`password=Juuso&email=foobar@mailinator.com`))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 404)
		})

		Convey("should return 302 with valid data", func() {
			request, _ := http.NewRequest("POST", "/user/login", strings.NewReader(fmt.Sprintf(`password=%s&email=%s`, user.Password, user.Email)))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 302)
		})
	})

	TestUserLogout(t)

	Convey("on JSON API", t, func() {
		var recorder = httptest.NewRecorder()

		Convey("should return 401 with wrong password", func() {
			request, _ := http.NewRequest("POST", "/api/user/login", strings.NewReader(`{"password": "Juuso", "email": "vertigo-test@mailinator.com"}`))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
		})

		Convey("should return 404 with wrong non-existent email", func() {
			request, _ := http.NewRequest("POST", "/api/user/login", strings.NewReader(`{"password": "Juuso", "email": "foobar@mailinator.com"}`))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 404)
		})

		Convey("should return 200 with valid data", func() {
			request, _ := http.NewRequest("POST", "/api/user/login", strings.NewReader(fmt.Sprintf(`{"password":"%s", "email":"%s"}`, user.Password, user.Email)))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			// i assure, nothing else worked
			sessioncookie = strings.Split(strings.TrimLeft(recorder.HeaderMap["Set-Cookie"][0], "id="), ";")[0]
		})
	})
}

func TestUserControlPanel(t *testing.T) {

	Convey("without authentication, it should fail", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/user", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 401)
	})

	Convey("with authentication, it should succeed", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/user", nil)
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
	})
}

func TestSessionRedirect(t *testing.T) {

	Convey("without authentication, it should not redirect", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/user/login", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
	})

	Convey("with authentication, it should redirect", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/user/login", nil)
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 302)
	})
}

func TestPostCreationPage(t *testing.T) {

	Convey("without authentication, it should 401", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/posts/new", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 401)
	})

	Convey("with authentication, it should 200 OK", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/posts/new", nil)
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
	})
}

func TestCreateFirstPost(t *testing.T) {
	testCreatePost(t, 1, "First post", "This is example post with HTML elements like **bold** and *italics* in place.")
}

func testCreatePostRequest(t *testing.T, payload []byte, p Post) {

	Convey("with authentication and valid data, it should 200 OK", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("POST", "/api/post", bytes.NewReader(payload))
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		json.Unmarshal(recorder.Body.Bytes(), &post)
		So(post.ID, ShouldEqual, p.ID)
		So(post.Title, ShouldEqual, p.Title)
		So(post.Content, ShouldEqual, p.Content)
		So(post.Markdown, ShouldEqual, p.Markdown)
		So(post.Slug, ShouldEqual, p.Slug)
		So(post.Author, ShouldEqual, p.Author)
		So(post.Created, ShouldBeGreaterThan, int64(1400000000))
		if post.Updated != post.Created {
			So(post.Updated, ShouldAlmostEqual, post.Created, 5)
		}
		So(post.Excerpt, ShouldEqual, p.Excerpt)
		So(post.Viewcount, ShouldEqual, p.Viewcount)
	})
}

func TestReadPost(t *testing.T) {

	Convey("using API", t, func() {

		Convey("with the latest post's slug, it should return 200 OK", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", fmt.Sprintf("/api/post/%s", post.Slug), nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			var p Post
			json.Unmarshal(recorder.Body.Bytes(), &p)
			So(post.ID, ShouldEqual, p.ID)
			So(post.Title, ShouldEqual, p.Title)
			So(post.Content, ShouldEqual, p.Content)
			So(post.Markdown, ShouldEqual, p.Markdown)
			So(post.Slug, ShouldEqual, p.Slug)
			So(post.Author, ShouldEqual, p.Author)
			So(post.Created, ShouldBeGreaterThan, int64(1400000000))
			if post.Updated != post.Created {
				So(post.Updated, ShouldAlmostEqual, post.Created, 5)
			}
			So(post.Excerpt, ShouldEqual, p.Excerpt)
			So(post.Viewcount, ShouldEqual, p.Viewcount)
			post.Viewcount += 1
			time.Sleep(1 * time.Second)
		})
	})

	Convey("using frontend", t, func() {

		Convey("with the latest post's slug, it should return 200 OK", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", fmt.Sprintf("/post/%s", post.Slug), nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			post.Viewcount += 1
			time.Sleep(1 * time.Second)
		})
	})
}

// func TestReadPostSpecialCases(t *testing.T) {

// 	Convey("should return error when accessing slug called `new`", t, func() {
// 		var recorder = httptest.NewRecorder()
// 		request, _ := http.NewRequest("GET", "/api/post/new", nil)
// 		server.ServeHTTP(recorder, request)
// 		So(recorder.Code, ShouldEqual, 400)
// 		So(recorder.Body.String(), ShouldEqual, `{"error":"There can't be a post called 'new'."}`)
// 	})
// }

func TestPublishPost(t *testing.T) {

	Convey("publishing post which does not exist", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/api/post/foobar/publish", nil)
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 404)
	})

	Convey("without session data should return HTTP 401", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", fmt.Sprintf("/api/post/%s/publish", post.Slug), nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 401)
	})

	Convey("with session data should return HTTP 200", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", fmt.Sprintf("/api/post/%s/publish", post.Slug), nil)
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		So(recorder.Body.String(), ShouldEqual, `{"success":"Post published"}`)
		So(recorder.Code, ShouldEqual, 200)
	})
}

func TestUnpublishPost(t *testing.T) {

	Convey("unpublishing post which does not exist", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/api/post/foobar/unpublish", nil)
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 404)
	})

	Convey("without session data should return HTTP 401", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", fmt.Sprintf("/api/post/%s/unpublish", post.Slug), nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 401)
	})

	Convey("with session data should return HTTP 200", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", fmt.Sprintf("/api/post/%s/unpublish", post.Slug), nil)
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		So(recorder.Body.String(), ShouldEqual, `{"success":"Post unpublished"}`)
	})

	Convey("after unpublishing, the post should be hidden", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/api/posts", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		So(recorder.Body.String(), ShouldEqual, `[]`)
	})

	TestPublishPost(t)
}

func TestPostListing(t *testing.T) {

	Convey("using API", t, func() {

		Convey("after a post is published, it should be listed on /api/posts", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/api/posts", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			var posts []Post
			json.Unmarshal(recorder.Body.Bytes(), &posts)
			for i, p := range posts {
				So(i, ShouldEqual, 0)
				So(post.ID, ShouldEqual, p.ID)
				So(post.Title, ShouldEqual, p.Title)
				So(post.Content, ShouldEqual, p.Content)
				So(post.Markdown, ShouldEqual, p.Markdown)
				So(post.Slug, ShouldEqual, p.Slug)
				So(post.Author, ShouldEqual, p.Author)
				So(post.Created, ShouldBeGreaterThan, int64(1400000000))
				if post.Updated != post.Created {
					So(post.Updated, ShouldAlmostEqual, post.Created, 5)
				}
				So(post.Excerpt, ShouldEqual, p.Excerpt)
				So(post.Viewcount, ShouldEqual, p.Viewcount)
			}
		})
	})

	Convey("using frontend", t, func() {

		Convey("after a post is published, it should be listed on frontpage", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			doc, _ := goquery.NewDocumentFromReader(recorder.Body)
			sel := doc.Find("article .title").Text()
			So(sel, ShouldEqual, post.Title)
		})
	})
}

func TestPostOwner(t *testing.T) {

	Convey("using API", t, func() {

		Convey("post owner should have data linked to their profile", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/api/user/1", nil)
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			var u User
			json.Unmarshal(recorder.Body.Bytes(), &u)
			So(u.ID, ShouldEqual, user.ID)
			So(u.Name, ShouldEqual, user.Name)
			So(u.Email, ShouldEqual, user.Email)
			p := u.Posts[0]
			So(post.ID, ShouldEqual, p.ID)
			So(post.Title, ShouldEqual, p.Title)
			So(post.Content, ShouldEqual, p.Content)
			So(post.Markdown, ShouldEqual, p.Markdown)
			So(post.Slug, ShouldEqual, p.Slug)
			So(post.Author, ShouldEqual, p.Author)
			So(post.Created, ShouldBeGreaterThan, int64(1400000000))
			if post.Updated != post.Created {
				So(post.Updated, ShouldAlmostEqual, post.Created, 5)
			}
			So(post.Excerpt, ShouldEqual, p.Excerpt)
			So(post.Viewcount, ShouldEqual, p.Viewcount)
		})
	})
}

func TestPostEditPage(t *testing.T) {

	Convey("should return 200 OK with authorization", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", fmt.Sprintf("/post/%s/edit", post.Slug), nil)
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
	})

	Convey("should return 401 without authorization", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", fmt.Sprintf("/post/%s/edit", post.Slug), nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 401)
	})
}

func TestUpdateFirstPost(t *testing.T) {
	testUpdatePost(t, "First post edited", "This is API edited example post with HTML elements like **bold** and *italics* in place.")
}

func testUpdatePostAPI(t *testing.T, payload []byte) {

	Convey("should return 401 without authorization", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("POST", fmt.Sprintf("/api/post/%s/edit", post.Slug), bytes.NewReader(payload))
		request.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 401)
		So(recorder.Body.String(), ShouldEqual, `{"error":"Unauthorized"}`)
	})

	Convey("should return 401 with bad authorization", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("POST", fmt.Sprintf("/api/post/%s/edit", post.Slug), bytes.NewReader(payload))
		cookie := &http.Cookie{Name: "id", Value: malformedsessioncookie}
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 401)
		So(recorder.Body.String(), ShouldEqual, `{"error":"Unauthorized"}`)
	})

	Convey("should return 404 with non-existent post", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("POST", "/api/post/foobar/edit", bytes.NewReader(payload))
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 404)
		So(recorder.Body.String(), ShouldEqual, `{"error":"Not found"}`)
	})

	Convey("should return 200 with successful authorization", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("POST", fmt.Sprintf("/api/post/%s/edit", post.Slug), bytes.NewReader(payload))
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		var p Post
		json.Unmarshal(recorder.Body.Bytes(), &p)
		post.Slug = p.Slug
		So(post.ID, ShouldEqual, p.ID)
		So(post.Title, ShouldEqual, p.Title)
		So(post.Content, ShouldEqual, p.Content)
		So(post.Markdown, ShouldEqual, p.Markdown)
		So(post.Slug, ShouldEqual, p.Slug)
		So(post.Author, ShouldEqual, p.Author)
		So(post.Created, ShouldBeGreaterThan, int64(1400000000))
		if post.Updated != post.Created {
			So(post.Updated, ShouldAlmostEqual, post.Created, 5)
		}
		So(post.Excerpt, ShouldEqual, p.Excerpt)
		So(post.Viewcount, ShouldEqual, p.Viewcount)
	})
}

func TestPostAfterUpdating(t *testing.T) {

	Convey("the post should not be displayed on frontpage", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		doc, _ := goquery.NewDocumentFromReader(recorder.Body)
		sel := doc.Find("article h1").Text()
		So(sel, ShouldBeEmpty)
	})

	Convey("update should return HTTP 200", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", fmt.Sprintf("/api/post/%s/publish", post.Slug), nil)
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		So(recorder.Body.String(), ShouldEqual, `{"success":"Post published"}`)
		So(recorder.Code, ShouldEqual, 200)
	})

	Convey("after updating, post should be displayed on frontpage", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		doc, _ := goquery.NewDocumentFromReader(recorder.Body)
		sel := doc.Find("article .title").Text()
		So(sel, ShouldEqual, post.Title)
	})

	Convey("the post should not be displayed trough API", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/api/posts", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		var posts []Post
		json.Unmarshal(recorder.Body.Bytes(), &posts)
		for i, p := range posts {
			So(i, ShouldEqual, 0)
			So(post.ID, ShouldEqual, p.ID)
			So(post.Title, ShouldEqual, p.Title)
			So(post.Content, ShouldEqual, p.Content)
			So(post.Markdown, ShouldEqual, p.Markdown)
			So(post.Slug, ShouldEqual, p.Slug)
			So(post.Author, ShouldEqual, p.Author)
			So(post.Created, ShouldBeGreaterThan, int64(1400000000))
			if post.Updated != post.Created {
				So(post.Updated, ShouldAlmostEqual, post.Created, 5)
			}
			So(post.Excerpt, ShouldEqual, p.Excerpt)
		}
	})
}

func testUpdatePostFrontend(t *testing.T, payload string) {

	Convey("should return 302 with successful authorization", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("POST", fmt.Sprintf("/post/%s/edit", post.Slug), strings.NewReader(payload))
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 302)
	})
}

func testCreatePost(t *testing.T, owner int64, title string, markdown string) {
	var u User
	u.ID = owner

	var p Post
	p.ID = post.ID + 1
	p.Title = title
	p.Markdown = markdown
	p.Content = string(blackfriday.MarkdownCommon([]byte(markdown)))
	p.Slug = slug.Create(p.Title)
	p.Author = u.ID
	p.Excerpt = excerpt.Make(p.Content, 15)
	p.Viewcount = 0
	payload, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	testCreatePostRequest(t, payload, p)
}

func TestCreateSecondPost(t *testing.T) {
	testCreatePost(t, 1, "Second post", "This is second post")
}

func testUpdatePost(t *testing.T, title string, markdown string) {
	// create JSON payload
	var p Post
	p.Title = title
	p.Markdown = markdown
	p.Content = string(blackfriday.MarkdownCommon([]byte(markdown)))
	apiPayload, _ := json.Marshal(p)

	// save changes to global object for further testing comparison
	post.Title = p.Title
	post.Content = string(blackfriday.MarkdownCommon([]byte(markdown)))
	post.Markdown = markdown
	post.Excerpt = excerpt.Make(p.Content, 15)

	testUpdatePostAPI(t, apiPayload)

	// creates form-encoded payload
	p2 := url.Values{}
	p2.Set("title", title)
	p2.Add("markdown", markdown)
	frontendPayload := p2.Encode()

	// save changes to global object for further testing comparison
	post.Title = p2.Get("title")
	post.Markdown = p2.Get("markdown")
	post.Excerpt = excerpt.Make(post.Content, 15)

	testUpdatePostFrontend(t, frontendPayload)
	TestReadPost(t)
}

func TestUpdateSecondPost(t *testing.T) {
	testUpdatePost(t, "Second post edited", "This is edited second post")
}

func TestControlPanelListing(t *testing.T) {

	Convey("should list both posts", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/user", nil)
		cookie := &http.Cookie{Name: "id", Value: sessioncookie}
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		doc, _ := goquery.NewDocumentFromReader(recorder.Body)
		doc.Find("ul").Each(func(i int, s *goquery.Selection) {
			So(i, ShouldBeLessThanOrEqualTo, post.ID)
		})
	})
}

func TestCreateThirdPost(t *testing.T) {
	testUpdatePost(t, "Third post", "This is third post")
	TestPublishPost(t)
	TestReadPost(t)
	TestControlPanelListing(t)
}

func TestDeletePost(t *testing.T) {

	Convey("using API", t, func() {

		Convey("it should return 401 without sessioncookies", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", fmt.Sprintf("/api/post/%s/delete", post.Slug), nil)
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Unauthorized"}`)
		})

		Convey("it should return 401 with malformed sessioncookie", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", fmt.Sprintf("/api/post/%s/delete", post.Slug), nil)
			cookie := &http.Cookie{Name: "id", Value: malformedsessioncookie}
			request.AddCookie(cookie)
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Unauthorized"}`)
		})

		Convey("it should return 404 when trying to delete non-existent post", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/api/post/foobar/delete", nil)
			cookie := &http.Cookie{Name: "id", Value: sessioncookie}
			request.AddCookie(cookie)
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 404)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Not found"}`)
		})

		Convey("it should return 200 with successful sessioncookies", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", fmt.Sprintf("/api/post/%s/delete", post.Slug), nil)
			cookie := &http.Cookie{Name: "id", Value: sessioncookie}
			request.AddCookie(cookie)
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			So(recorder.Body.String(), ShouldEqual, `{"success":"Post deleted"}`)
			if *Driver == "sqlite3" {
				// SQLite's re-assigns ID if one is removed
				post.ID--
			}
		})
	})

	TestControlPanelListing(t)
}

func TestReadSettings(t *testing.T) {

	Convey("using API", t, func() {

		Convey("it should return 401 without sessioncookies", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/api/settings", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Unauthorized"}`)
		})

		Convey("reading with malformed sessioncookies it should return 401", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/api/settings", nil)
			cookie := &http.Cookie{Name: "id", Value: malformedsessioncookie}
			request.AddCookie(cookie)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Unauthorized"}`)
		})

		Convey("reading with sessioncookies it should return 200", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/api/settings", nil)
			cookie := &http.Cookie{Name: "id", Value: sessioncookie}
			request.AddCookie(cookie)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)

			var returnedSettings Vertigo
			json.Unmarshal(recorder.Body.Bytes(), &returnedSettings)

			var settingsWithoutCookieHash = settings
			settingsWithoutCookieHash.CookieHash = ""

			So(returnedSettings, ShouldResemble, settingsWithoutCookieHash)
		})
	})

	Convey("using frontend", t, func() {

		Convey("without sessioncookies it should return 401", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/user/settings", nil)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
		})

		Convey("with malformed sessioncookies it should return 401", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/user/settings", nil)
			cookie := &http.Cookie{Name: "id", Value: malformedsessioncookie}
			request.AddCookie(cookie)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
		})

		Convey("with successful sessioncookies it should return 200", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/user/settings", nil)
			cookie := &http.Cookie{Name: "id", Value: sessioncookie}
			request.AddCookie(cookie)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
		})
	})

}

func TestUpdateSettings(t *testing.T) {

	var s Vertigo
	s = settings
	s.Name = "Foo's Blog"
	s.AllowRegistrations = false
	payload, _ := json.Marshal(s)

	testUpdateSettings(t, payload, s)

	settings.Name = s.Name
	settings.AllowRegistrations = false

	TestReadSettings(t)
}

func testUpdateSettings(t *testing.T, payload []byte, s Vertigo) {

	Convey("using API", t, func() {

		Convey("without sessioncookies", func() {
			var recorder = httptest.NewRecorder()
			payload, _ := json.Marshal(settings)
			request, _ := http.NewRequest("POST", "/api/settings", bytes.NewReader(payload))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Unauthorized"}`)
		})

		Convey("with successful sessioncookies", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/api/settings", bytes.NewReader(payload))
			cookie := &http.Cookie{Name: "id", Value: sessioncookie}
			request.AddCookie(cookie)
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			So(recorder.Body.String(), ShouldEqual, `{"success":"Settings were successfully saved"}`)
			So(Settings, ShouldResemble, &s)
		})
	})
}

func TestAllowRegistration(t *testing.T) {

	Convey("creation after allowregistrations is false", t, func() {

		Convey("should return HTTP 403", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/api/user", strings.NewReader(`{"name": "Juuso", "password": "hello", "email": "bar@example.com"}`))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 403)
			So(recorder.Body.String(), ShouldEqual, `{"error":"New registrations are not allowed at this time."}`)
		})

		Convey("should return HTTP 403 on frontend", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/user/register", strings.NewReader(`name=Juuso&password=hello&email=bar@example.com`))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 403)
		})
	})
}

func TestMarkdown(t *testing.T) {

	var s Vertigo
	s = settings
	payload, _ := json.Marshal(s)

	// switch settings to Markdown
	testUpdateSettings(t, payload, s)

	TestReadSettings(t)
	TestControlPanelListing(t)
	TestPostCreationPage(t)

	post.Markdown = "### foo\n*foo* foo **foo**"
	post.Content = string(blackfriday.MarkdownCommon([]byte(post.Markdown)))

	testCreatePost(t, 1, "Markdown post", post.Markdown)
	TestPublishPost(t)
}

func TestFeeds(t *testing.T) {

	Convey("reading feeds", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/rss", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		So(recorder.HeaderMap["Content-Type"][0], ShouldEqual, "application/xml")
	})
}

func TestSearch(t *testing.T) {

	Convey("using API", t, func() {

		Convey("searching for the latest post should return it", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/api/posts/search", strings.NewReader(fmt.Sprintf(`{"query": "%s"}`, "Markdown")))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			var posts []Post
			json.Unmarshal(recorder.Body.Bytes(), &posts)
			for i, p := range posts {
				So(i, ShouldEqual, 0)
				So(post.ID, ShouldEqual, p.ID)
				So(post.Title, ShouldEqual, p.Title)
				So(post.Content, ShouldEqual, p.Content)
				So(post.Markdown, ShouldEqual, p.Markdown)
				So(post.Slug, ShouldEqual, p.Slug)
				So(post.Author, ShouldEqual, p.Author)
				So(post.Created, ShouldBeGreaterThan, int64(1400000000))
				if post.Updated != post.Created {
					So(post.Updated, ShouldAlmostEqual, post.Created, 5)
				}
				So(post.Excerpt, ShouldEqual, p.Excerpt)
				So(post.Viewcount, ShouldEqual, p.Viewcount)
			}
		})

		Convey("searching for non-existent post should return empty JSON array", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/api/posts/search", strings.NewReader(`{"query": "fizzbar"}`))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			So(recorder.Body.String(), ShouldEqual, "[]")
		})
	})

	Convey("on frontend", t, func() {

		Convey("searching for the latest post using title", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/posts/search", strings.NewReader(fmt.Sprintf(`query=%s`, "Markdown")))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			doc, _ := goquery.NewDocumentFromReader(recorder.Body)
			sel := doc.Find(".title").Text()
			So(sel, ShouldEqual, post.Title)
		})

		Convey("searching for the latest post using content", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/posts/search", strings.NewReader(`query=foo`))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			doc, _ := goquery.NewDocumentFromReader(recorder.Body)
			sel := doc.Find(".title").Text()
			So(sel, ShouldEqual, post.Title)
		})

		Convey("searching with a query which is not contained in any post", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/posts/search", strings.NewReader(`query=foofoobarbar`))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			doc, _ := goquery.NewDocumentFromReader(recorder.Body)
			sel := doc.Find("h2").Text()
			So(sel, ShouldEqual, "Nothing found.")
		})
	})
}

func TestUserLogout(t *testing.T) {

	Convey("using API", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/api/user/logout", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 200)
		So(recorder.Body.String(), ShouldEqual, `{"success":"You've been logged out."}`)
	})

	Convey("using frontend", t, func() {
		var recorder = httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/user/logout", nil)
		server.ServeHTTP(recorder, request)
		So(recorder.Code, ShouldEqual, 302)
	})
}

func TestPasswordRecovery(t *testing.T) {

	Convey("using frontend", t, func() {

		Convey("should return 401 with email which does not exist", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/user/recover", strings.NewReader(`email=foobar@example.com`))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
			So(recorder.Body.String(), ShouldEqual, `{"error":"User with that email does not exist."}`)
		})

		Convey("should return 302 with latest user email", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/user/recover", strings.NewReader(fmt.Sprintf(`email=%s`, user.Email)))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 302)
		})
	})

	Convey("using API", t, func() {

		Convey("should return 200 with latest user email", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/api/user/recover", strings.NewReader(fmt.Sprintf(`{"email": "%s"}`, user.Email)))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			So(recorder.Body.String(), ShouldEqual, `{"success":"We've sent you a link to your email which you may use you reset your password."}`)
		})
	})

	testShouldRecoveryFieldBeBlank(t, false)
}

func testShouldRecoveryFieldBeBlank(t *testing.T, value bool) {

	Convey("the latest user should have recovery key defined", t, func() {
		user, _ = user.GetByEmail()
		if value == false {
			So(strings.TrimSpace(user.Recovery), ShouldNotBeBlank)
		} else {
			So(strings.TrimSpace(user.Recovery), ShouldEqual, "")
		}
		recovery = user.Recovery
	})
}

func TestPasswordReset(t *testing.T) {

	Convey("using frontend", t, func() {

		Convey("should return 400 when ID is malformed", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/user/reset/foobar/"+recovery, strings.NewReader(`password=newpassword`))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 400)
			So(recorder.Body.String(), ShouldEqual, `{"error":"User ID could not be parsed from request URL."}`)
		})

		Convey("should return 400 when recovery UUID does not pass UUID checks", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/user/reset/1/foobar", strings.NewReader(`password=newpassword`))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 400)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Could not parse UUID from the request."}`)
		})

		Convey("should return 400 when user with given ID does not exist", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", fmt.Sprintf(`/user/reset/%d/%s`, 7, recovery), strings.NewReader(`password=newpassword`))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 400)
			So(recorder.Body.String(), ShouldEqual, `{"error":"User with that ID does not exist."}`)
		})
	})

	Convey("using API", t, func() {

		Convey("should return 200 with valid information", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", fmt.Sprintf(`/api/user/reset/%d/%s`, user.ID, recovery), strings.NewReader(`{"password":"newpassword"}`))
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 200)
			user.Password = "newpassword"
		})
	})

	testShouldRecoveryFieldBeBlank(t, true)
	TestUserSignin(t)
}

func TestRecoveryKeyExpiration(t *testing.T) {

	Convey("using frontend", t, func() {

		Convey("should redirect to login page with notification", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/user/recover", strings.NewReader(fmt.Sprintf(`email=%s`, user.Email)))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 302)
		})
	})

	testShouldRecoveryFieldBeBlank(t, false)

	Convey("recovery key should be blank after expiring", t, func() {

		var u User
		u.ID = int64(1)
		go u.ExpireRecovery(1 * time.Second)
		time.Sleep(2 * time.Second)

		u, _ = user.Get()
		So(strings.TrimSpace(u.Recovery), ShouldEqual, "")
	})
}

func TestPostSecurity(t *testing.T) {

	var s Vertigo
	s = settings
	s.AllowRegistrations = true
	payload, _ := json.Marshal(s)

	testUpdateSettings(t, payload, s)

	settings.AllowRegistrations = true

	TestReadSettings(t)

	user.Name = "Juuso"
	user.Password = "foo"
	user.Email = "vertigo-test2@mailinator.com"
	user.Location = "Europe/Helsinki"

	testCreateUser(t, user.Name, user.Password, user.Email, user.Location)
	TestUserSignin(t)

	Convey("using API", t, func() {

		Convey("updating post of another user", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/api/post/"+post.Slug+"/edit", strings.NewReader(`{"title": "First post edited twice", "markdown": "This is an EDITED example post with HTML elements like **bold** and *italics* in place."}`))
			cookie := &http.Cookie{Name: "id", Value: sessioncookie}
			request.AddCookie(cookie)
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Unauthorized"}`)
		})

		Convey("publishing post of another user", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/api/post/"+post.Slug+"/publish", nil)
			cookie := &http.Cookie{Name: "id", Value: sessioncookie}
			request.AddCookie(cookie)
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Unauthorized"}`)
		})

		Convey("deleting post of another user", func() {
			var recorder = httptest.NewRecorder()
			request, _ := http.NewRequest("GET", "/api/post/"+post.Slug+"/delete", nil)
			cookie := &http.Cookie{Name: "id", Value: sessioncookie}
			request.AddCookie(cookie)
			request.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(recorder, request)
			So(recorder.Code, ShouldEqual, 401)
			So(recorder.Body.String(), ShouldEqual, `{"error":"Unauthorized"}`)
		})
	})
}

func TestDropDatabase(t *testing.T) {
	Drop()
}
