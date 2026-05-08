package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/omar/sentinel-proxy/internal/config"
	"github.com/omar/sentinel-proxy/internal/metrics"
	"github.com/omar/sentinel-proxy/internal/middleware"
)

type App struct {
	Config *config.Config
}

// =========================
// SSE CLIENT STORAGE
// =========================

var (
	clients   = make(map[chan string]bool)
	clientsMu sync.Mutex
)

// NewApp initializes the App struct required by main.go
func NewApp() *App {
	return &App{
		Config: config.Load(),
	}
}

// =========================
// SSE BROADCASTER
// =========================

func broadcast(message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		select {
		case client <- message:
		default:
		}
	}
}

// =========================
// LOG STREAM ENDPOINT
// =========================

func logsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	clientChan := make(chan string)

	clientsMu.Lock()
	clients[clientChan] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, clientChan)
		clientsMu.Unlock()
		close(clientChan)
	}()

	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	notify := r.Context().Done()

	for {
		select {
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-notify:
			return
		}
	}
}

// =========================
// IDP PROXY
// =========================

func proxyTo(target *url.URL, w http.ResponseWriter, r *http.Request) {
	targetAddr := target.String() + r.URL.Path

	if r.URL.RawQuery != "" {
		targetAddr += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, targetAddr, r.Body)
	if err != nil {
		log.Printf("Failed to create proxy request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header = r.Header
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("IDP (Backend) Unreachable: %v", err)
		http.Error(w, "IDP Unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// DecodeUsernameFromToken pulls 'bob' out of the JWT
func DecodeUsernameFromToken(tokenString string) (string, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		if name, ok := claims["username"].(string); ok {
			return name, nil
		}

		if sub, ok := claims["sub"].(string); ok {
			return sub, nil
		}
	}

	return "", fmt.Errorf("invalid token")
}

// =========================
// SERVER START
// =========================

func (a *App) Start() {
	idpRaw := os.Getenv("IDP_URL")
	if idpRaw == "" {
		idpRaw = "http://idp:8080"
	}

	idpURL, err := url.Parse(idpRaw)
	if err != nil {
		log.Fatal("Invalid IDP_URL:", err)
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(idpURL)
	mux := http.NewServeMux()
	idpProxy := httputil.NewSingleHostReverseProxy(idpURL)

	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpProxy.ServeHTTP(w, r)
	})

	mux.Handle("/auth/", authHandler)
	mux.Handle("/login/", authHandler)
	mux.Handle("/register/", authHandler)
	mux.Handle("/", authHandler)

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics.GetStats())
	})
	mux.HandleFunc("/logs", logsHandler)

	// MAIN HANDLER WITH MIDDLEWARE
	finalHandler := middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Identify the user FIRST
		userID := "anonymous"

		auth := r.Header.Get("Authorization")

		// Fallback to access_token cookie
		if auth == "" {
			cookie, err := r.Cookie("access_token")
			if err == nil {
				auth = cookie.Value
			}
		}

		// Remove Bearer prefix if present
		if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
			auth = after
		}

		log.Printf("AUTH TOKEN: %s", auth)

		if auth != "" {
			if name, err := DecodeUsernameFromToken(auth); err == nil {
				userID = name
			}
		}

		log.Printf("DECODED USER: %s", userID)

		// Attach the Identity to the Request Context
		// update 'r' directly so that all subsequent handlers see the user_id
		ctx := context.WithValue(r.Context(), "user_id", userID)
		r = r.WithContext(ctx)

		securedHandler := middleware.Chain(
			middleware.RequestID,
			middleware.RateLimiter,
			middleware.WAF, // WAF will now find "jon" in the context
		)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This block handles cases where the WAF flags but allows the request to continue[cite: 6]

			reverseProxy.ServeHTTP(w, r)
		}))

		// Execute the chain with the updated request 'r'
		securedHandler.ServeHTTP(w, r)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Sentinel Proxy started on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, finalHandler))
}
