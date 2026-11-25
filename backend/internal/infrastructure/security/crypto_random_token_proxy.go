package security

import (
	"crypto/rand"
	"encoding/base64"
)

// randomTokenBytes is 256 bits: unguessable, and enough that a plain SHA-256 digest is a safe way to store it.
const randomTokenBytes = 32

// CryptoRandomTokenProxy draws secrets from the operating system's cryptographic random source.
type CryptoRandomTokenProxy struct{}

func NewCryptoRandomTokenProxy() *CryptoRandomTokenProxy {
	return &CryptoRandomTokenProxy{}
}

// NewToken relies on crypto/rand.Read, which never returns an error (it crashes the program instead).
func (tokenProxy *CryptoRandomTokenProxy) NewToken() string {
	randomBytes := make([]byte, randomTokenBytes)
	_, _ = rand.Read(randomBytes)

	return base64.RawURLEncoding.EncodeToString(randomBytes)
}
