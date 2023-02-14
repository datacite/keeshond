package auth

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/datacite/keeshond/internal/app"
	"github.com/go-chi/jwtauth/v5"
)

func GetAuthToken(config *app.Config) *jwtauth.JWTAuth {
	publicKeyBlock, _ := pem.Decode([]byte(config.DataCite.JWTPublicKey))

	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)

	if err != nil {
		fmt.Println("Error parsing public key")
	}

	// Private key is nil because we are only using the public key to verify the token.
	return jwtauth.New("RS256", nil, publicKey)
}