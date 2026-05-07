package crypto

import (
	"crypto/rand"
	"crypto/rsa"
)

// GenerateIdPKeys creates a 2048-bit RSA key pair for signing JWTs
func GenerateIdPKeys() (*rsa.PrivateKey, error) {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return nil, err
	}

	return key, nil
}
