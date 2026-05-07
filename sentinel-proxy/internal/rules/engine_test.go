package rules

import (
	"net/http"
	"net/url"
	"testing"
)

func TestSQLInjectionDetection(t *testing.T) {
	LoadRules()
	req := &http.Request{
		URL: &url.URL{
			Path:     "/",
			RawQuery: "id=1 union select",
		},
	}

	blocked, reason := EvaluateRequest(req, req.URL.RawQuery)

	if !blocked {
		t.Errorf("Expected request to be blocked for SQLi")
	}

	if reason == "" {
		t.Errorf("Expected a rule name, got empty")
	}
}

func TestXSSDetection(t *testing.T) {
	LoadRules()
	req := &http.Request{
		URL: &url.URL{
			Path:     "/",
			RawQuery: "q=<script>alert(1)</script>",
		},
	}

	blocked, _ := EvaluateRequest(req, req.URL.RawQuery)

	if !blocked {
		t.Errorf("Expected XSS to be blocked")
	}
}

func TestPathBlocking(t *testing.T) {
	LoadRules()
	req := &http.Request{
		URL: &url.URL{
			Path: "/admin",
		},
	}

	blocked, _ := EvaluateRequest(req, "")

	if !blocked {
		t.Errorf("Expected /admin to be blocked")
	}
}

func TestCleanRequest(t *testing.T) {
	LoadRules()
	req := &http.Request{
		URL: &url.URL{
			Path:     "/home",
			RawQuery: "id=123",
		},
	}

	blocked, _ := EvaluateRequest(req, req.URL.RawQuery)

	if blocked {
		t.Errorf("Expected clean request to pass")
	}
}