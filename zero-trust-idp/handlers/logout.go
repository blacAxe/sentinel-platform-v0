package handlers

import (
	"net/http"

	"github.com/omar/zero-trust-idp/db"
)

func Logout(w http.ResponseWriter, r *http.Request) {

	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Missing refresh token", http.StatusBadRequest)
		return
	}

	refreshToken := cookie.Value

	// Hash it
	hashed := HashToken(refreshToken)

	// Delete session from DB
	err = db.DeleteSession(hashed)
	if err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	// Clear access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	// Respond
	w.Write([]byte("Logged out successfully"))
}