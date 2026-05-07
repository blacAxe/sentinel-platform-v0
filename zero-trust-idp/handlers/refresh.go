package handlers

import (
	"net/http"

	"github.com/omar/zero-trust-idp/db"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from header
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Missing refresh token", http.StatusBadRequest)
		return
	}

	refreshToken := cookie.Value

	// Hash it 
	hashed := HashToken(refreshToken)

	// Check DB for valid session
	userID, err := db.GetSession(hashed)
	if err != nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// Issue new access token
	newAccessToken, err := GenerateAccessToken(userID, "user")
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   900,
	})

	// Return new access token
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{
		"access_token": "` + newAccessToken + `"
	}`))
}
