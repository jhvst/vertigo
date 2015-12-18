// Copyright 2015 Husobee Associates, LLC.  All rights reserved.
// Use of this source code is governed by The MIT License, which
// can be found in the LICENSE file included.
// Portions Copyright 2013 Julien Schmidt.  All rights reserved.
// Use of portions of this source code is governed by a BSD style License
// which can be found in the LICENSE-go-http-routing-benchmark file included.
// Portions Copyright 2015 Labstack.  All rights reserved.

package vestigo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type route struct {
	method  string
	path    string
	handler http.HandlerFunc
}

var (
	api = []route{
		// OAuth Authorizations
		{"GET", "/authorizations", nil},
		{"GET", "/authorizations/:id", nil},
		{"POST", "/authorizations", nil},
		{"PUT", "/authorizations/clients/:client_id", nil},
		{"PATCH", "/authorizations/:id", nil},
		{"DELETE", "/authorizations/:id", nil},
		{"GET", "/applications/:client_id/tokens/:access_token", nil},
		{"DELETE", "/applications/:client_id/tokens", nil},
		{"DELETE", "/applications/:client_id/tokens/:access_token", nil},

		// Activity
		{"GET", "/events", nil},
		{"GET", "/repos/:owner/:repo/events", nil},
		{"GET", "/networks/:owner/:repo/events", nil},
		{"GET", "/orgs/:org/events", nil},
		{"GET", "/users/:user/received_events", nil},
		{"GET", "/users/:user/received_events/public", nil},
		{"GET", "/users/:user/events", nil},
		{"GET", "/users/:user/events/public", nil},
		{"GET", "/users/:user/events/orgs/:org", nil},
		{"GET", "/feeds", nil},
		{"GET", "/notifications", nil},
		{"GET", "/repos/:owner/:repo/notifications", nil},
		{"PUT", "/notifications", nil},
		{"PUT", "/repos/:owner/:repo/notifications", nil},
		{"GET", "/notifications/threads/:id", nil},
		{"PATCH", "/notifications/threads/:id", nil},
		{"GET", "/notifications/threads/:id/subscription", nil},
		{"PUT", "/notifications/threads/:id/subscription", nil},
		{"DELETE", "/notifications/threads/:id/subscription", nil},
		{"GET", "/repos/:owner/:repo/stargazers", nil},
		{"GET", "/users/:user/starred", nil},
		{"GET", "/user/starred", nil},
		{"GET", "/user/starred/:owner/:repo", nil},
		{"PUT", "/user/starred/:owner/:repo", nil},
		{"DELETE", "/user/starred/:owner/:repo", nil},
		{"GET", "/repos/:owner/:repo/subscribers", nil},
		{"GET", "/users/:user/subscriptions", nil},
		{"GET", "/user/subscriptions", nil},
		{"GET", "/repos/:owner/:repo/subscription", nil},
		{"PUT", "/repos/:owner/:repo/subscription", nil},
		{"DELETE", "/repos/:owner/:repo/subscription", nil},
		{"GET", "/user/subscriptions/:owner/:repo", nil},
		{"PUT", "/user/subscriptions/:owner/:repo", nil},
		{"DELETE", "/user/subscriptions/:owner/:repo", nil},

		// Gists
		{"GET", "/users/:user/gists", nil},
		{"GET", "/gists", nil},
		{"GET", "/gists/public", nil},
		{"GET", "/gists/starred", nil},
		{"GET", "/gists/:id", nil},
		{"POST", "/gists", nil},
		{"PATCH", "/gists/:id", nil},
		{"PUT", "/gists/:id/star", nil},
		{"DELETE", "/gists/:id/star", nil},
		{"GET", "/gists/:id/star", nil},
		{"POST", "/gists/:id/forks", nil},
		{"DELETE", "/gists/:id", nil},

		// Git Data
		{"GET", "/repos/:owner/:repo/git/blobs/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/blobs", nil},
		{"GET", "/repos/:owner/:repo/git/commits/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/commits", nil},
		{"GET", "/repos/:owner/:repo/git/refs/*ref", nil},
		{"GET", "/repos/:owner/:repo/git/refs", nil},
		{"POST", "/repos/:owner/:repo/git/refs", nil},
		{"PATCH", "/repos/:owner/:repo/git/refs/*ref", nil},
		{"DELETE", "/repos/:owner/:repo/git/refs/*ref", nil},
		{"GET", "/repos/:owner/:repo/git/tags/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/tags", nil},
		{"GET", "/repos/:owner/:repo/git/trees/:sha", nil},
		{"POST", "/repos/:owner/:repo/git/trees", nil},

		// Issues
		{"GET", "/issues", nil},
		{"GET", "/user/issues", nil},
		{"GET", "/orgs/:org/issues", nil},
		{"GET", "/repos/:owner/:repo/issues", nil},
		{"GET", "/repos/:owner/:repo/issues/:number", nil},
		{"POST", "/repos/:owner/:repo/issues", nil},
		{"PATCH", "/repos/:owner/:repo/issues/:number", nil},
		{"GET", "/repos/:owner/:repo/assignees", nil},
		{"GET", "/repos/:owner/:repo/assignees/:assignee", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/comments", nil},
		{"GET", "/repos/:owner/:repo/issues/comments", nil},
		{"GET", "/repos/:owner/:repo/issues/comments/:id", nil},
		{"POST", "/repos/:owner/:repo/issues/:number/comments", nil},
		{"PATCH", "/repos/:owner/:repo/issues/comments/:id", nil},
		{"DELETE", "/repos/:owner/:repo/issues/comments/:id", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/events", nil},
		{"GET", "/repos/:owner/:repo/issues/events", nil},
		{"GET", "/repos/:owner/:repo/issues/events/:id", nil},
		{"GET", "/repos/:owner/:repo/labels", nil},
		{"GET", "/repos/:owner/:repo/labels/:name", nil},
		{"POST", "/repos/:owner/:repo/labels", nil},
		{"PATCH", "/repos/:owner/:repo/labels/:name", nil},
		{"DELETE", "/repos/:owner/:repo/labels/:name", nil},
		{"GET", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"POST", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name", nil},
		{"PUT", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"DELETE", "/repos/:owner/:repo/issues/:number/labels", nil},
		{"GET", "/repos/:owner/:repo/milestones/:number/labels", nil},
		{"GET", "/repos/:owner/:repo/milestones", nil},
		{"GET", "/repos/:owner/:repo/milestones/:number", nil},
		{"POST", "/repos/:owner/:repo/milestones", nil},
		{"PATCH", "/repos/:owner/:repo/milestones/:number", nil},
		{"DELETE", "/repos/:owner/:repo/milestones/:number", nil},

		// Miscellaneous
		{"GET", "/emojis", nil},
		{"GET", "/gitignore/templates", nil},
		{"GET", "/gitignore/templates/:name", nil},
		{"POST", "/markdown", nil},
		{"POST", "/markdown/raw", nil},
		{"GET", "/meta", nil},
		{"GET", "/rate_limit", nil},

		// Organizations
		{"GET", "/users/:user/orgs", nil},
		{"GET", "/user/orgs", nil},
		{"GET", "/orgs/:org", nil},
		{"PATCH", "/orgs/:org", nil},
		{"GET", "/orgs/:org/members", nil},
		{"GET", "/orgs/:org/members/:user", nil},
		{"DELETE", "/orgs/:org/members/:user", nil},
		{"GET", "/orgs/:org/public_members", nil},
		{"GET", "/orgs/:org/public_members/:user", nil},
		{"PUT", "/orgs/:org/public_members/:user", nil},
		{"DELETE", "/orgs/:org/public_members/:user", nil},
		{"GET", "/orgs/:org/teams", nil},
		{"GET", "/teams/:id", nil},
		{"POST", "/orgs/:org/teams", nil},
		{"PATCH", "/teams/:id", nil},
		{"DELETE", "/teams/:id", nil},
		{"GET", "/teams/:id/members", nil},
		{"GET", "/teams/:id/members/:user", nil},
		{"PUT", "/teams/:id/members/:user", nil},
		{"DELETE", "/teams/:id/members/:user", nil},
		{"GET", "/teams/:id/repos", nil},
		{"GET", "/teams/:id/repos/:owner/:repo", nil},
		{"PUT", "/teams/:id/repos/:owner/:repo", nil},
		{"DELETE", "/teams/:id/repos/:owner/:repo", nil},
		{"GET", "/user/teams", nil},

		// Pull Requests
		{"GET", "/repos/:owner/:repo/pulls", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number", nil},
		{"POST", "/repos/:owner/:repo/pulls", nil},
		{"PATCH", "/repos/:owner/:repo/pulls/:number", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/commits", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/files", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/merge", nil},
		{"PUT", "/repos/:owner/:repo/pulls/:number/merge", nil},
		{"GET", "/repos/:owner/:repo/pulls/:number/comments", nil},
		{"GET", "/repos/:owner/:repo/pulls/comments", nil},
		{"GET", "/repos/:owner/:repo/pulls/comments/:number", nil},
		{"PUT", "/repos/:owner/:repo/pulls/:number/comments", nil},
		{"PATCH", "/repos/:owner/:repo/pulls/comments/:number", nil},
		{"DELETE", "/repos/:owner/:repo/pulls/comments/:number", nil},

		// Repositories
		{"GET", "/user/repos", nil},
		{"GET", "/users/:user/repos", nil},
		{"GET", "/orgs/:org/repos", nil},
		{"GET", "/repositories", nil},
		{"POST", "/user/repos", nil},
		{"POST", "/orgs/:org/repos", nil},
		{"GET", "/repos/:owner/:repo", nil},
		{"PATCH", "/repos/:owner/:repo", nil},
		{"GET", "/repos/:owner/:repo/contributors", nil},
		{"GET", "/repos/:owner/:repo/languages", nil},
		{"GET", "/repos/:owner/:repo/teams", nil},
		{"GET", "/repos/:owner/:repo/tags", nil},
		{"GET", "/repos/:owner/:repo/branches", nil},
		{"GET", "/repos/:owner/:repo/branches/:branch", nil},
		{"DELETE", "/repos/:owner/:repo", nil},
		{"GET", "/repos/:owner/:repo/collaborators", nil},
		{"GET", "/repos/:owner/:repo/collaborators/:user", nil},
		{"PUT", "/repos/:owner/:repo/collaborators/:user", nil},
		{"DELETE", "/repos/:owner/:repo/collaborators/:user", nil},
		{"GET", "/repos/:owner/:repo/comments", nil},
		{"GET", "/repos/:owner/:repo/commits/:sha/comments", nil},
		{"POST", "/repos/:owner/:repo/commits/:sha/comments", nil},
		{"GET", "/repos/:owner/:repo/comments/:id", nil},
		{"PATCH", "/repos/:owner/:repo/comments/:id", nil},
		{"DELETE", "/repos/:owner/:repo/comments/:id", nil},
		{"GET", "/repos/:owner/:repo/commits", nil},
		{"GET", "/repos/:owner/:repo/commits/:sha", nil},
		{"GET", "/repos/:owner/:repo/readme", nil},
		{"GET", "/repos/:owner/:repo/contents/*path", nil},
		{"PUT", "/repos/:owner/:repo/contents/*path", nil},
		{"DELETE", "/repos/:owner/:repo/contents/*path", nil},
		{"GET", "/repos/:owner/:repo/:archive_format/:ref", nil},
		{"GET", "/repos/:owner/:repo/keys", nil},
		{"GET", "/repos/:owner/:repo/keys/:id", nil},
		{"POST", "/repos/:owner/:repo/keys", nil},
		{"PATCH", "/repos/:owner/:repo/keys/:id", nil},
		{"DELETE", "/repos/:owner/:repo/keys/:id", nil},
		{"GET", "/repos/:owner/:repo/downloads", nil},
		{"GET", "/repos/:owner/:repo/downloads/:id", nil},
		{"DELETE", "/repos/:owner/:repo/downloads/:id", nil},
		{"GET", "/repos/:owner/:repo/forks", nil},
		{"POST", "/repos/:owner/:repo/forks", nil},
		{"GET", "/repos/:owner/:repo/hooks", nil},
		{"GET", "/repos/:owner/:repo/hooks/:id", nil},
		{"POST", "/repos/:owner/:repo/hooks", nil},
		{"PATCH", "/repos/:owner/:repo/hooks/:id", nil},
		{"POST", "/repos/:owner/:repo/hooks/:id/tests", nil},
		{"DELETE", "/repos/:owner/:repo/hooks/:id", nil},
		{"POST", "/repos/:owner/:repo/merges", nil},
		{"GET", "/repos/:owner/:repo/releases", nil},
		{"GET", "/repos/:owner/:repo/releases/:id", nil},
		{"POST", "/repos/:owner/:repo/releases", nil},
		{"PATCH", "/repos/:owner/:repo/releases/:id", nil},
		{"DELETE", "/repos/:owner/:repo/releases/:id", nil},
		{"GET", "/repos/:owner/:repo/releases/:id/assets", nil},
		{"GET", "/repos/:owner/:repo/stats/contributors", nil},
		{"GET", "/repos/:owner/:repo/stats/commit_activity", nil},
		{"GET", "/repos/:owner/:repo/stats/code_frequency", nil},
		{"GET", "/repos/:owner/:repo/stats/participation", nil},
		{"GET", "/repos/:owner/:repo/stats/punch_card", nil},
		{"GET", "/repos/:owner/:repo/statuses/:ref", nil},
		{"POST", "/repos/:owner/:repo/statuses/:ref", nil},

		// Search
		{"GET", "/search/repositories", nil},
		{"GET", "/search/code", nil},
		{"GET", "/search/issues", nil},
		{"GET", "/search/users", nil},
		{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword", nil},
		{"GET", "/legacy/repos/search/:keyword", nil},
		{"GET", "/legacy/user/search/:keyword", nil},
		{"GET", "/legacy/user/email/:email", nil},

		// Users
		{"GET", "/users/:user", nil},
		{"GET", "/user", nil},
		{"PATCH", "/user", nil},
		{"GET", "/users", nil},
		{"GET", "/user/emails", nil},
		{"POST", "/user/emails", nil},
		{"DELETE", "/user/emails", nil},
		{"GET", "/users/:user/followers", nil},
		{"GET", "/user/followers", nil},
		{"GET", "/users/:user/following", nil},
		{"GET", "/user/following", nil},
		{"GET", "/user/following/:user", nil},
		{"GET", "/users/:user/following/:target_user", nil},
		{"PUT", "/user/following/:user", nil},
		{"DELETE", "/user/following/:user", nil},
		{"GET", "/users/:user/keys", nil},
		{"GET", "/user/keys", nil},
		{"GET", "/user/keys/:id", nil},
		{"POST", "/user/keys", nil},
		{"PATCH", "/user/keys/:id", nil},
		{"DELETE", "/user/keys/:id", nil},
	}
)

func TestRouterParam(t *testing.T) {
	r := NewRouter()
	r.Add("GET", "/users/:id", func(w http.ResponseWriter, r *http.Request) {})
	req, _ := http.NewRequest("GET", "/users/1", nil)
	w := httptest.NewRecorder()
	h := r.Find(req)
	if assert.NotNil(t, h) {
		h(w, req)
		assert.Equal(t, "1", Param(req, "id"))
	}
}

func TestRouterTwoParam(t *testing.T) {
	r := NewRouter()
	r.Add("GET", "/users/:uid/files/:fid", func(w http.ResponseWriter, r *http.Request) {})

	req, _ := http.NewRequest("GET", "/users/1/files/1", nil)
	w := httptest.NewRecorder()
	h := r.Find(req)
	if assert.NotNil(t, h) {
		h(w, req)
		assert.Equal(t, "1", Param(req, "uid"))
		assert.Equal(t, "1", Param(req, "fid"))
	}
}

/*
func TestRouterMatchAny(t *testing.T) {
	r := NewRouter()
	r.Add(GET, "/users/*", func(w http.ResponseWriter, r *http.Request) {})

	req, _ := http.NewRequest("GET", "/users/", nil)

	h, _ := r.Find(req)
	if assert.NotNil(t, h) {
		assert.Equal(t, "", c.P(0))
	}

	req2, _ := http.NewRequest("GET", "/users/1", nil)

	h, _ = r.Find(req2)
	if assert.NotNil(t, h) {
		assert.Equal(t, "1", c.P(0))
	}
}
*/

func TestRouterMicroParam(t *testing.T) {
	r := NewRouter()
	r.Add("GET", "/:a/:b/:c", func(w http.ResponseWriter, r *http.Request) {})

	req, _ := http.NewRequest("GET", "/1/2/3", nil)
	h := r.Find(req)
	if assert.NotNil(t, h) {
		assert.Equal(t, "1", Param(req, "a"))
		assert.Equal(t, "2", Param(req, "b"))
		assert.Equal(t, "3", Param(req, "c"))
	}
}

func TestRouterMixParamMatchAny(t *testing.T) {
	r := NewRouter()

	// Route
	r.Add("GET", "/users/:id/*", func(w http.ResponseWriter, r *http.Request) {})

	req, _ := http.NewRequest("GET", "/users/joe/comments", nil)
	w := httptest.NewRecorder()
	h := r.Find(req)
	if assert.NotNil(t, h) {
		h(w, req)
		assert.Equal(t, "joe", Param(req, "id"))
	}
}

func TestRouterMultiRoute(t *testing.T) {
	r := NewRouter()

	// Routes
	r.Add("GET", "/users", func(w http.ResponseWriter, r *http.Request) {})
	r.Add("GET", "/users/:id", func(w http.ResponseWriter, r *http.Request) {})

	// Route > /users
	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	h := r.Find(req)
	if assert.NotNil(t, h) {
		h(w, req)
	}

	// Route > /users/:id
	req, _ = http.NewRequest("GET", "/users/1", nil)
	h = r.Find(req)
	w = httptest.NewRecorder()
	if assert.NotNil(t, h) {
		h(w, req)
		assert.Equal(t, "1", Param(req, "id"))
	}

	// Route > /user
	req, _ = http.NewRequest("GET", "/user", nil)
	h = r.Find(req)
	w = httptest.NewRecorder()

	h(w, req)
	assert.Equal(t, w.Code, http.StatusNotFound)

	// Invalid Method for Resource
	// Route > /user
	req, _ = http.NewRequest("INVALID", "/users", nil)
	h = r.Find(req)
	w = httptest.NewRecorder()
	h(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

}

/*
func TestRouterPriority(t *testing.T) {
	r := NewRouter()

	// Routes
	r.Add(GET, "/users", func(w http.ResponseWriter, r *http.Request) {})
	r.Add(GET, "/users/new", func(w http.ResponseWriter, r *http.Request) {})
	r.Add(GET, "/users/:id", func(w http.ResponseWriter, r *http.Request) {})
	r.Add(GET, "/users/dew", func(w http.ResponseWriter, r *http.Request) {})
	r.Add(GET, "/users/:id/files", func(w http.ResponseWriter, r *http.Request) {})
	r.Add(GET, "/users/newsee", func(w http.ResponseWriter, r *http.Request) {})
	r.Add(GET, "/users/*", func(w http.ResponseWriter, r *http.Request) {})

	// Route > /users
	h, _ := r.Find(GET, "/users", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 1, c.Get("a"))
	}

	// Route > /users/new
	h, _ = r.Find(GET, "/users/new", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 2, c.Get("b"))
	}

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/1", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 3, c.Get("c"))
	}

	// Route > /users/dew
	h, _ = r.Find(GET, "/users/dew", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 4, c.Get("d"))
	}

	// Route > /users/:id/files
	h, _ = r.Find(GET, "/users/1/files", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 5, c.Get("e"))
	}

	// Route > /users/:id
	h, _ = r.Find(GET, "/users/news", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 3, c.Get("c"))
	}

	// Route > /users/*
	h, _ = r.Find(GET, "/users/joe/books", c)
	if assert.NotNil(t, h) {
		h(c)
		assert.Equal(t, 7, c.Get("g"))
		assert.Equal(t, "joe/books", c.Param("_name"))
	}
}
*/

func TestRouterParamNames(t *testing.T) {
	r := NewRouter()

	// Routes
	r.Add("GET", "/users", func(w http.ResponseWriter, r *http.Request) {})
	r.Add("GET", "/users/:uid", func(w http.ResponseWriter, r *http.Request) {})
	r.Add("GET", "/users/:uid/files/:fid", func(w http.ResponseWriter, r *http.Request) {})

	// Route > /users/:id
	req, _ := http.NewRequest("GET", "/users/1", nil)
	w := httptest.NewRecorder()
	h := r.Find(req)
	if assert.NotNil(t, h) {
		h(w, req)
		assert.Equal(t, "1", Param(req, "uid"))
	}

	// Route > /users/:uid/files/:fid
	req, _ = http.NewRequest("GET", "/users/1/files/1", nil)
	w = httptest.NewRecorder()
	h = r.Find(req)
	r.root.printTree("", true)
	if assert.NotNil(t, h) {
		h(w, req)
		fmt.Println(req.Form)
		fmt.Println(ParamNames(req))

		assert.Equal(t, "1", Param(req, "uid"))
		assert.Equal(t, "1", Param(req, "fid"))
	}
}

func TestRouterAPI(t *testing.T) {
	r := NewRouter()

	for _, route := range api {
		r.Add(route.method, route.path, func(w http.ResponseWriter, req *http.Request) {})
	}

	w := httptest.NewRecorder()

	for _, route := range api {

		req, _ := http.NewRequest(route.method, route.path, nil)
		h := r.Find(req)
		if assert.NotNil(t, h) {
			for _, n := range ParamNames(req) {
				if assert.NotEmpty(t, n) {
					assert.NotNil(t, Param(req, n))
				}
			}
			h(w, req)
		}
	}
}

func TestRouterAddInvalidMethod(t *testing.T) {
	r := NewRouter()
	assert.Panics(t, func() {
		r.Add("INVALID", "/", func(w http.ResponseWriter, req *http.Request) {})
	})
}

func TestRouterServeHTTP(t *testing.T) {
	r := NewRouter()

	r.Add("GET", "/users", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(""))
	})

	// OK
	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Not found
	req, _ = http.NewRequest("GET", "/files", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
