package proxy

import (
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
	"time"

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

// shipToLumen sends security events to the Lumen Ingestor
func shipToLumen(r *http.Request, action string, attackType string) {
	ingestorURL := "http://lumen-ingestor:9001/events"

	// --- IDENTITY EXTRACTION ---
	userID := "anonymous"
	auth := r.Header.Get("Authorization")
	if auth == "" {
		if cookie, err := r.Cookie("access_token"); err == nil {
			auth = cookie.Value
		}
	}

	if auth != "" {
		if name, err := DecodeUsernameFromToken(auth); err == nil {
			userID = name
		}
	}

	event := map[string]string{
		"user_id":     userID,
		"attack_type": attackType,
		"action":      action,
	}

	if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
		event["request_id"] = reqID
	}

	payload, _ := json.Marshal(event)
	client := &http.Client{Timeout: 2 * time.Second}

	go func() {
		resp, err := client.Post(ingestorURL, "application/json", strings.NewReader(string(payload)))
		if err == nil {
			resp.Body.Close()
		}
	}()
}

// DecodeUsernameFromToken pulls 'bob' out of the JWT
func DecodeUsernameFromToken(tokenString string) (string, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	// Parse without validation because the IDP already validated it; we just need the name
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "anonymous", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if username, ok := claims["sub"].(string); ok {
			return username, nil
		}
	}
	return "anonymous", fmt.Errorf("no sub claim found")
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

	// IDP Routes
	idpProxy := httputil.NewSingleHostReverseProxy(idpURL)

	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpProxy.ServeHTTP(w, r)
	})

	mux.Handle("/auth/", authHandler)
	mux.Handle("/login/", authHandler)
	mux.Handle("/register/", authHandler)
	mux.Handle("/", authHandler)

	// Sentinel Platform / Dashboard Routes
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics.GetStats())
	})
	mux.HandleFunc("/logs", logsHandler)

	// MAIN HANDLER WITH MIDDLEWARE
	finalHandler := middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("REQUEST -> %s %s", r.Method, r.URL.Path)
		path := r.URL.Path

		// Let Sentinel Platform and IDP routes bypass the heavy WAF/Limiter chain
		if strings.HasPrefix(path, "/stats") ||
			strings.HasPrefix(path, "/logs") {
			mux.ServeHTTP(w, r)
			return
		}

		// Security Chain for the main application
		securedHandler := middleware.Chain(
			middleware.RequestID,
			middleware.RateLimiter,
			middleware.WAF,
		)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			attack := r.Header.Get("X-Sentinel-Attack")

			if attack != "" {

				shipToLumen(r, "blocked", attack)

				userID := "anonymous"

				auth := r.Header.Get("Authorization")

				if auth == "" {
					if cookie, err := r.Cookie("access_token"); err == nil {
						auth = cookie.Value
					}
				}

				if auth != "" {
					if name, err := DecodeUsernameFromToken(auth); err == nil {
						userID = name
					}
				}

				event := map[string]interface{}{
					"user_id":     userID,
					"attack_type": attack,
					"action":      "blocked",
					"method":      r.Method,
					"path":        r.URL.Path,
					"ip":          r.RemoteAddr,
					"timestamp":   time.Now().Unix(),
				}

				payload, _ := json.Marshal(event)
				broadcast(string(payload))
			}

			log.Printf("Forwarding to IDP: %s", r.URL.Path)

			reverseProxy.ServeHTTP(w, r)
		}))

		securedHandler.ServeHTTP(w, r)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Sentinel Proxy started on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, finalHandler))
}
