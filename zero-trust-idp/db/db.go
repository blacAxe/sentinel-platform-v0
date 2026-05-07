package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	_ "github.com/lib/pq"
)

var DB *sql.DB

// Initialize DB connection
func InitDB() error {
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		return fmt.Errorf("DB_URL not set")
	}

	log.Printf("Connecting to DB with string: %s", connStr)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}

		log.Println("Waiting for DB to be ready...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("could not connect to DB after retries: %w", err)
	}

	DB = db
	return nil
}

type User struct {
	ID          string
	Name        string
	Credentials []webauthn.Credential
}

// WebAuthn interface methods
func (u *User) WebAuthnID() []byte                         { return []byte(u.ID) }
func (u *User) WebAuthnName() string                       { return u.Name }
func (u *User) WebAuthnDisplayName() string                { return u.Name }
func (u *User) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }
func (u *User) WebAuthnIcon() string                       { return "" }

func GetUser(username string) (*User, error) {
	query := `SELECT id, username, credentials FROM users WHERE username=$1`

	var id int64
	var credsJSON string

	err := DB.QueryRow(query, username).Scan(&id, &username, &credsJSON)
	if err != nil {
		return nil, err
	}

	var creds []webauthn.Credential

	if len(credsJSON) > 0 {
		err := json.Unmarshal([]byte(credsJSON), &creds)
		if err != nil {
			return nil, err
		}
	}

	log.Println("LOADED CREDS:", len(creds))

	return &User{
		ID:          fmt.Sprintf("%d", id),
		Name:        username,
		Credentials: creds,
	}, nil
}

func CreateUser(username string) (*User, error) {
	query := `INSERT INTO users (username, credentials, created_at)
	          VALUES ($1, $2, $3)
	          RETURNING id`

	emptyCreds, _ := json.Marshal([]webauthn.Credential{})

	var id int64
	err := DB.QueryRow(query, username, emptyCreds, time.Now()).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:          fmt.Sprintf("%d", id),
		Name:        username,
		Credentials: []webauthn.Credential{},
	}, nil
}

func SaveUser(u *User) error {
	log.Println("SAVING USER WITH CREDS:", len(u.Credentials))

	credsJSON, _ := json.Marshal(u.Credentials)

	query := `UPDATE users SET credentials=$1 WHERE username=$2`
	_, err := DB.Exec(query, string(credsJSON), u.Name)
	return err
}

func (u *User) AddCredential(cred webauthn.Credential) {
	u.Credentials = append(u.Credentials, cred)
}

func CreateSession(userID int, hashedToken string, expiry time.Time) error {
	query := `INSERT INTO sessions (user_id, refresh_token_hash, expires_at, created_at)
              VALUES ($1, $2, $3, $4)`

	_, err := DB.Exec(query, userID, hashedToken, expiry, time.Now())
	return err
}

func GetSession(hashedToken string) (string, error) {
	query := `SELECT user_id FROM sessions WHERE refresh_token=$1 AND expires_at > NOW()`

	var userID string
	err := DB.QueryRow(query, hashedToken).Scan(&userID)
	return userID, err
}

func DeleteSession(hashedToken string) error {
	query := `DELETE FROM sessions WHERE refresh_token=$1`
	_, err := DB.Exec(query, hashedToken)
	return err
}
