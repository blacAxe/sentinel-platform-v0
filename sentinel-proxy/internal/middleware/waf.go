package middleware

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/omar/sentinel-proxy/internal/events"
	"github.com/omar/sentinel-proxy/internal/logger"
	"github.com/omar/sentinel-proxy/internal/metrics"
	"github.com/omar/sentinel-proxy/internal/rules"
)

func WAF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, _ := r.Context().Value(RequestIDKey).(string)
		decodedQuery, _ := url.QueryUnescape(r.URL.RawQuery)
		query := strings.ToLower(decodedQuery)

		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
		}

		metrics.IncTotal()
		blocked, reason := rules.EvaluateRequest(r, query)

		if blocked {
			// Pull the ID we passed from proxy.go
			val := r.Context().Value("user_id")
			userID, _ := val.(string)
			if userID == "" {
				userID = "anonymous"
			}

			fmt.Printf("WAF USER: %s\n", userID)

			event := events.SecurityEvent{
				EventType:      "request_blocked",
				RequestID:      requestID,
				User:           userID, 
				IP:             ip,
				Path:           r.URL.Path,
				Method:         r.Method,
				Query:          r.URL.RawQuery,
				AttackDetected: true,
				AttackType:     reason,
				Action:         "blocked",
				Timestamp:      time.Now().Unix(),
			}

			logger.LogEvent(event)
			events.SendEvent(event)
			metrics.IncBlocked()

			fmt.Printf("DEBUG: WAF context user_id is: %v\n", r.Context().Value("user_id"))

			// NOW it is safe to block the user
			http.Error(w, "Blocked by Sentinel", http.StatusForbidden)
			return
		}

		// Log Allowed for Terminal visualization

		next.ServeHTTP(w, r)
	})
}
