package strict

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAccepts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Martini Strict tests")
}

var _ = Describe("Negotiator", func() {
	var n Negotiator
	var r *http.Request
	BeforeEach(func() {
		var err error
		r, err = http.NewRequest("POST", "http://example.com/", nil)
		Expect(err).
			NotTo(HaveOccured())
		n = &negotiator{r}
	})
	It("should parse the Accept header correctly", func() {
		r.Header.Set("Accept", "application/json,text/xml;q=0.8")
		Expect(n.Accepts("application/json")).
			To(Equal(1.0))
		Expect(n.Accepts("text/xml")).
			To(Equal(0.8))
		Expect(n.Accepts("text/csv")).
			To(Equal(0.0))
	})
	It("should parse the Content-Type header correctly", func() {
		r.Header.Set("Content-Type", "application/json; charset=UTF-8")
		Expect(n.ContentType("application/json")).
			To(BeTrue())
		Expect(n.ContentType("text/plain")).
			To(BeFalse())
	})
})

var _ = Describe("ContentType", func() {
	var w *httptest.ResponseRecorder
	var r *http.Request
	BeforeEach(func() {
		var err error
		r, err = http.NewRequest("POST", "http://example.com/", nil)
		Expect(err).
			NotTo(HaveOccured())
		w = httptest.NewRecorder()
	})
	It("should accept requests with a matching content type", func() {
		r.Header.Set("Content-Type", "application/json")
		ContentType("application/json")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusUnsupportedMediaType))
	})
	It("should accept requests with a matching content type with extra values", func() {
		r.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ContentType("application/json")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusUnsupportedMediaType))
	})
	It("should accept requests with a matching content type when multiple content types are supported", func() {
		r.Header.Set("Content-Type", "text/xml; charset=UTF-8")
		ContentType("application/json", "text/xml")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusUnsupportedMediaType))
	})
	It("should accept requests with no content type if empty content type headers are allowed", func() {
		ContentType("application/json", "text/xml", "")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusUnsupportedMediaType))
	})
	It("should not accept requests with no content type if empty content type headers are not allowed", func() {
		ContentType("application/json", "text/xml")(w, r)
		Expect(w.Code).
			To(Equal(http.StatusUnsupportedMediaType))
	})
	It("should not accept requests with a mismatching content type", func() {
		r.Header.Set("Content-Type", "text/plain")
		ContentType("application/json", "text/xml")(w, r)
		Expect(w.Code).
			To(Equal(http.StatusUnsupportedMediaType))
	})
	It("should not accept requests with a mismatching content type even if empty content types are allowed", func() {
		r.Header.Set("Content-Type", "text/plain")
		ContentType("application/json", "text/xml", "")(w, r)
		Expect(w.Code).
			To(Equal(http.StatusUnsupportedMediaType))
	})
	It("should act on block POST, PATCH and PUT requests", func() {
		var err error
		for _, m := range []string{"POST", "PATCH", "PUT"} {
			r, err = http.NewRequest(m, "http://example.com/", nil)
			Expect(err).
				NotTo(HaveOccured())
			r.Header.Set("Content-Type", "text/plain")
			ContentType("application/json", "text/xml", "")(w, r)
			Expect(w.Code).
				To(Equal(http.StatusUnsupportedMediaType))
		}
	})
	It("should not block GET, HEAD, OPTIONS and DELETE requests", func() {
		var err error
		for _, m := range []string{"GET", "HEAD", "OPTIONS", "DELETE"} {
			r, err = http.NewRequest(m, "http://example.com/", nil)
			Expect(err).
				NotTo(HaveOccured())
			r.Header.Set("Content-Type", "text/plain")
			ContentType("application/json", "text/xml", "")(w, r)
			Expect(w.Code).
				NotTo(Equal(http.StatusUnsupportedMediaType))
		}
	})
})

var _ = Describe("ContentCharset", func() {
	var w *httptest.ResponseRecorder
	var r *http.Request
	BeforeEach(func() {
		var err error
		r, err = http.NewRequest("POST", "http://example.com/", nil)
		Expect(err).
			NotTo(HaveOccured())
		w = httptest.NewRecorder()
	})
	It("should accept requests with a matching charset", func() {
		r.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ContentCharset("UTF-8")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusUnsupportedMediaType))
	})
	It("should be case-insensitive", func() {
		r.Header.Set("Content-Type", "application/json; charset=utf-8")
		ContentCharset("UTF-8")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusUnsupportedMediaType))
	})
	It("should accept requests with a matching charset with extra values", func() {
		r.Header.Set("Content-Type", "application/json; foo=bar; charset=UTF-8; spam=eggs")
		ContentCharset("UTF-8")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusUnsupportedMediaType))
	})
	It("should accept requests with a matching charset when multiple charsets are supported", func() {
		r.Header.Set("Content-Type", "text/xml; charset=UTF-8")
		ContentCharset("UTF-8", "Latin-1")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusUnsupportedMediaType))
	})
	It("should accept requests with no charset if empty charset headers are allowed", func() {
		r.Header.Set("Content-Type", "text/xml")
		ContentCharset("UTF-8", "")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusUnsupportedMediaType))
	})
	It("should not accept requests with no charset if empty charset headers are not allowed", func() {
		r.Header.Set("Content-Type", "text/xml")
		ContentCharset("UTF-8")(w, r)
		Expect(w.Code).
			To(Equal(http.StatusUnsupportedMediaType))
	})
	It("should not accept requests with a mismatching charset", func() {
		r.Header.Set("Content-Type", "text/plain; charset=Latin-1")
		ContentCharset("UTF-8")(w, r)
		Expect(w.Code).
			To(Equal(http.StatusUnsupportedMediaType))
	})
	It("should not accept requests with a mismatching charset even if empty charsets are allowed", func() {
		r.Header.Set("Content-Type", "text/plain; charset=Latin-1")
		ContentCharset("UTF-8", "")(w, r)
		Expect(w.Code).
			To(Equal(http.StatusUnsupportedMediaType))
	})
	It("should act on block POST, PATCH and PUT requests", func() {
		var err error
		for _, m := range []string{"POST", "PATCH", "PUT"} {
			r, err = http.NewRequest(m, "http://example.com/", nil)
			Expect(err).
				NotTo(HaveOccured())
			r.Header.Set("Content-Type", "text/plain")
			ContentType("application/json", "text/xml", "")(w, r)
			Expect(w.Code).
				To(Equal(http.StatusUnsupportedMediaType))
		}
	})
	It("should not block GET, HEAD, OPTIONS and DELETE requests", func() {
		var err error
		for _, m := range []string{"GET", "HEAD", "OPTIONS", "DELETE"} {
			r, err = http.NewRequest(m, "http://example.com/", nil)
			Expect(err).
				NotTo(HaveOccured())
			r.Header.Set("Content-Type", "text/plain")
			ContentType("application/json", "text/xml", "")(w, r)
			Expect(w.Code).
				NotTo(Equal(http.StatusUnsupportedMediaType))
		}
	})
})

var _ = Describe("Accept", func() {
	var w *httptest.ResponseRecorder
	var r *http.Request
	BeforeEach(func() {
		var err error
		r, err = http.NewRequest("POST", "http://example.com/", nil)
		Expect(err).
			NotTo(HaveOccured())
		w = httptest.NewRecorder()
	})
	It("should accept requests with a matching content type", func() {
		r.Header.Set("Accept", "application/json")
		Accept("application/json")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusNotAcceptable))
	})
	It("should accept requests with a matching content type when multiple content types are supported", func() {
		r.Header.Set("Accept", "text/xml")
		Accept("application/json", "text/xml")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusNotAcceptable))
	})
	It("should accept requests with a matching content type when multiple content types are acceptable", func() {
		r.Header.Set("Accept", "text/xml,application/json")
		Accept("application/json")(w, r)
		Expect(w.Code).
			ToNot(Equal(http.StatusNotAcceptable))
	})
	It("should not accept requests when no matching pairs are found", func() {
		r.Header.Set("Accept", "image/webp,image/png")
		Accept("application/json", "text/xml")(w, r)
		Expect(w.Code).
			To(Equal(http.StatusNotAcceptable))
	})
})

var _ = Describe("accepts", func() {
	It("should return the correct q value", func() {
		a := "text/html,application/xhtml+xml;q=0.9,image/webp,image/*;q=0.8;,*/*;q=0.6"
		Expect(accepts(a, "text/html")).
			To(Equal(1.0))
		Expect(accepts(a, "image/webp")).
			To(Equal(1.0))
		Expect(accepts(a, "application/xhtml+xml")).
			To(Equal(0.9))
		Expect(accepts(a, "image/png")).
			To(Equal(0.8))
		Expect(accepts(a, "text/csv")).
			To(Equal(0.6))
	})
	It("should return the correct q value even if not acceptable", func() {
		a := "text/html,application/json;level=2;q=0.2"
		Expect(accepts(a, "text/html")).
			To(Equal(1.0))
		Expect(accepts(a, "application/json")).
			To(Equal(0.2))
		Expect(accepts(a, "image/png")).
			To(Equal(0.0))
	})
	It("should return the correct q value when everything is acceptable", func() {
		a := ""
		Expect(accepts(a, "text/html")).
			To(Equal(1.0))
		Expect(accepts(a, "application/json")).
			To(Equal(1.0))
		Expect(accepts(a, "image/png")).
			To(Equal(1.0))
	})
})
