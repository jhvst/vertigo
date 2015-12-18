package vestigo_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/husobee/vestigo"
)

func ExampleManyRoutes() {
	// new router
	router := vestigo.NewRouter()
	// standard http.HandlerFunc
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("version: %s, resource: %s\n", vestigo.Param(r, "version"), r.URL.Path)
	}
	// setup a GET /v:version/hi endpoint route in router
	router.Get("/v:version/hi", handler)
	// setup a GET /v:version/hi endpoint route in router
	router.Post("/v:version/hi", handler)
	// setup a GET /v:version/hi endpoint route in router
	router.Put("/v:version/hi", handler)
	// setup a GET /v:version/hi endpoint route in router
	router.Delete("/v:version/hi", handler)
	// setup a GET /v:version/hi endpoint route in router
	router.Patch("/v:version/hi", handler)

	// create a new request and response writer
	r, _ := http.NewRequest("PATCH", "/v2.3/hi", nil)
	w := httptest.NewRecorder()

	// execute the request
	router.ServeHTTP(w, r)
	// Output: version: 2.3, resource: /v2.3/hi
}

func ExampleSimpleRoute() {
	// new router
	router := vestigo.NewRouter()
	// standard http.HandlerFunc
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("version: %s, resource: %s\n", vestigo.Param(r, "version"), r.URL.Path)
	}
	// setup a GET /v:version/hi endpoint route in router
	router.Get("/v:version/hi", handler)

	// create a new request and response writer
	r, _ := http.NewRequest("GET", "/v2.3/hi", nil)
	w := httptest.NewRecorder()

	// execute the request
	router.ServeHTTP(w, r)
	// Output: version: 2.3, resource: /v2.3/hi
}

func ExampleCorsRoute() {
	// new router
	router := vestigo.NewRouter()
	// setup global cors config for router
	router.SetGlobalCors(&vestigo.CorsAccessControl{
		AllowOrigin:      []string{"*", "test.com"},          // allow these origins
		AllowCredentials: true,                               // credentials is allowed globally
		ExposeHeaders:    []string{"X-Header", "X-Y-Header"}, // Expose these headers
		MaxAge:           3600 * time.Second,                 // Cache max age
		AllowHeaders:     []string{"X-Header", "X-Y-Header"}, // Allow these headers
	})

	// standard http.HandlerFunc
	handler := func(w http.ResponseWriter, r *http.Request) {}

	// setup a GET/v:version/hi endpoint route in router
	router.Get("/v:version/hi", handler)

	// Setup a CORS policy for a particular route
	router.SetCors("/v:version/hi", &vestigo.CorsAccessControl{
		AllowMethods: []string{"HEAD"},
		AllowHeaders: []string{"X-Header", "X-Z-Header"},
	})

	// create a new request and response writer
	r, _ := http.NewRequest("OPTIONS", "/v2.3/hi", nil)
	// Initiate CORS
	r.Header.Add("Origin", "test.com")
	r.Header.Add("Access-Control-Request-Method", "HEAD")

	w := httptest.NewRecorder()
	// execute the request
	router.ServeHTTP(w, r)

	fmt.Println(w.Header().Get("Access-Control-Allow-Methods"))
	// Output: HEAD
}
