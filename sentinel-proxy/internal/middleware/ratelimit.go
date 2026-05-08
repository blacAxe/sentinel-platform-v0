package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/omar/sentinel-proxy/internal/events"
	"github.com/omar/sentinel-proxy/internal/logger"
	"github.com/omar/sentinel-proxy/internal/metrics"
)

type Client struct {
	Requests []int64
}

var clients = make(map[string]*Client)

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Path

		if path == "/favicon.ico" ||
			strings.Contains(path, ".well-known") {
			next.ServeHTTP(w, r)
			return
		}

		if strings.Contains(r.URL.Path, "favicon.ico") ||
			strings.Contains(r.URL.Path, ".well-known") {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/stats") ||
			strings.HasPrefix(r.URL.Path, "/logs") ||
			strings.HasPrefix(r.URL.Path, "/dashboard") {
			next.ServeHTTP(w, r)
			return
		}

		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
		}

		now := time.Now().Unix()

		client, exists := clients[ip]
		if !exists {
			client = &Client{}
			clients[ip] = client
		}

		var valid []int64
		for _, t := range client.Requests {
			if now-t < 10 {
				valid = append(valid, t)
			}
		}

		client.Requests = append(valid, now)

		if len(client.Requests) > 10 {
			requestID := r.Context().Value(RequestIDKey).(string)

			val := r.Context().Value("user_id")
			userID, _ := val.(string)

			if userID == "" {
				userID = "anonymous"
			}

			event := events.SecurityEvent{
				EventType:      "rate_limited",
				RequestID:      requestID,
				User:           userID,
				IP:             ip,
				Path:           r.URL.Path,
				Method:         r.Method,
				Query:          r.URL.RawQuery,
				AttackDetected: true,
				AttackType:     "RATE_LIMIT",
				Action:         "blocked",
				Timestamp:      time.Now().Unix(),
			}

			logger.LogEvent(event)
			events.SendEvent(event)
			metrics.IncBlocked()
			metrics.IncAttack("RATE_LIMIT")
			// ---------------------

			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
