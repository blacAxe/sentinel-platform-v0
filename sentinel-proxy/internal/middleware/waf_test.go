package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/omar/sentinel-proxy/internal/rules"
)

func TestWAFBlocksSQLi(t *testing.T) {
	rules.SetRules([]rules.Rule{
		{
			Name:    "SQL Injection",
			Pattern: "union select",
			Field:   "query",
		},
	})

	req := httptest.NewRequest("GET", "/?id=1%20union%20select", nil)
	w := httptest.NewRecorder()

	handler := Chain(
		RequestID,
		WAF,
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", w.Code)
	}
}

func TestWAFAllowsCleanRequest(t *testing.T) {
	rules.SetRules([]rules.Rule{
		{
			Name:    "SQL Injection",
			Pattern: "union select",
			Field:   "query",
		},
	})

	req := httptest.NewRequest("GET", "/?id=123", nil)
	w := httptest.NewRecorder()

	handler := Chain(
		RequestID,
		WAF,
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}