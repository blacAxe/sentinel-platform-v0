package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(username, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	log.Println("SIGNING WITH SECRET:", secret)

	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"iss":      "zero-trust-idp",
		"sub":      username,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			fmt.Println("JWT Error:", err)
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			if claims.Role != "admin" {
				fmt.Println("Access Denied for role:", claims.Role)
				http.Error(w, "Forbidden: Admins only!", http.StatusForbidden)
				return
			}
			next(w, r)
		} else {
			http.Error(w, "Invalid Claims", http.StatusUnauthorized)
		}
	}
}
