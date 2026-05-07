package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/omar/zero-trust-idp/db"
)

func BeginLogin(w http.ResponseWriter, r *http.Request, wa *webauthn.WebAuthn) {
	username := r.URL.Query().Get("username")
	user, err := db.GetUser(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// For Login, use wa.BeginLogin instead of BeginRegistration
	options, sessionData, err := wa.BeginLogin(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionDataStore[username] = sessionData
	JSONResponse(w, options)
}

func FinishLogin(w http.ResponseWriter, r *http.Request, wa *webauthn.WebAuthn) {
	log.Println("HIT /login/finish")

	username := r.URL.Query().Get("username")

	// Get the user from DB package
	user, err := db.GetUser(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// Get the session data stored in LoginBegin
	sessionData, ok := sessionDataStore[username]
	if !ok {
		http.Error(w, "Session not found", http.StatusBadRequest)
		return
	}

	log.Println("USER CREDS COUNT:", len(user.Credentials))

	// IMPORTANT: Actually verify the WebAuthn response from the browser
	// This checks if the hardware signature is valid
	_, err = wa.FinishLogin(user, *sessionData, r)
	if err != nil {
		log.Println("LOGIN ERROR:", err)
		http.Error(w, "Failed to verify passkey: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Determine the role
	role := "user"
	if user.WebAuthnName() == "bob" {
		role = "admin"
	}

	// Generate access token (short-lived)
	log.Println("IDP JWT SECRET:", os.Getenv("JWT_SECRET"))

	accessToken, err := GenerateAccessToken(user.WebAuthnName(), role)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	log.Println("ACCESS TOKEN:", accessToken)

	// Generate refresh token (long-lived)
	refreshToken := GenerateRefreshToken()

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   900, // 15 mins
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   604800, // 7 days
	})

	// Hash refresh token before storing
	hashedToken := HashToken(refreshToken)

	uid, _ := strconv.Atoi(user.ID)

	// Store session in DB
	err = db.CreateSession(
		uid,
		hashedToken,
		time.Now().Add(7*24*time.Hour),
	)

	if err != nil {
		http.Error(w, "Failed to store session", http.StatusInternalServerError)
		return
	}

	// 5. Send both tokens back
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{
		"status": "success",
		"access_token": "` + accessToken + `",
		"refresh_token": "` + refreshToken + `"
	}`))
}
