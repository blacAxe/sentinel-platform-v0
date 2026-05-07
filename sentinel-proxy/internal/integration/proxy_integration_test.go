package integration

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/omar/sentinel-proxy/internal/middleware"
	"github.com/omar/sentinel-proxy/internal/rules"
)

func setupTestProxy(targetURL string) http.Handler {

	rules.LoadRules()

	target, _ := url.Parse(targetURL)

	proxy := httputil.NewSingleHostReverseProxy(target)

	chain := middleware.Chain(
		middleware.RequestID,
		middleware.RateLimiter,
		middleware.WAF,
	)

	return chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}))
}

func TestAllowedRequest(t *testing.T) {

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("allowed"))
	}))
	defer backend.Close()

	handler := setupTestProxy(backend.URL)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}

func TestSQLInjectionBlocked(t *testing.T) {

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	handler := setupTestProxy(backend.URL)

	req := httptest.NewRequest(
		"GET",
		"/?id=1%20UNION%20SELECT%20password%20FROM%20users",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusOK {
		t.Fatalf("expected request to be blocked")
	}
}
