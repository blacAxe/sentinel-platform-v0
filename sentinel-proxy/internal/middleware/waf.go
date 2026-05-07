package middleware

import (
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


			event := events.SecurityEvent{
				EventType:      "request_blocked",
				RequestID:      requestID,
				User:           "anonymous",
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

			// --- ADD THIS LINE ---
			shipToLumen(event)
			// ---------------------

			http.Error(w, "Blocked by Sentinel", http.StatusForbidden)
			return
		}

		// Log Allowed for Terminal visualization

		next.ServeHTTP(w, r)
	})
}
