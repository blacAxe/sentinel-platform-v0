package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/omar/zero-trust-idp/db"
)

// In-memory store for session data
var sessionDataStore = make(map[string]*webauthn.SessionData)

func BeginRegistration(w http.ResponseWriter, r *http.Request, wa *webauthn.WebAuthn) {
	
	log.Println("HIT /register/begin") // 🔥 ADD THIS
	
	query := r.URL.Query()
	username := query.Get("username")
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	// Find or Create User
	user, err := db.GetUser(username)
	if err != nil {
		log.Println("GET USER FAILED:", err)

		user, err = db.CreateUser(username)
		if err != nil {
			log.Println("CREATE USER FAILED:", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		log.Println("USER CREATED:", username)
	} else {
		log.Println("USER FOUND:", username)
	}

	// Generate Registration Options
	options, sessionData, err := wa.BeginRegistration(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store session data to verify against in the Finish step
	sessionDataStore[username] = sessionData

	// Send options to frontend
	w.Header().Set("Content-Type", "application/json")
	log.Println("SENDING REGISTRATION OPTIONS")
	JSONResponse(w, options)
}

func FinishRegistration(w http.ResponseWriter, r *http.Request, wa *webauthn.WebAuthn) {
	username := r.URL.Query().Get("username")

	// Get the stored session data
	sessionData, exists := sessionDataStore[username]
	if !exists {
		http.Error(w, "Session not found", http.StatusBadRequest)
		return
	}

	user, err := db.GetUser(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// Parse the credential response from the browser
	credential, err := wa.FinishRegistration(user, *sessionData, r)
	if err != nil {
		http.Error(w, "Failed to verify: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Save the Public Key to the user's account
	user.AddCredential(*credential)
	err = db.SaveUser(user)
	if err != nil {
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Registration Successful! Passkey saved."))
}

func JSONResponse(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(data)
}
