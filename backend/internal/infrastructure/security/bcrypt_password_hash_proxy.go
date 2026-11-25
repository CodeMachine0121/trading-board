package security

import "golang.org/x/crypto/bcrypt"

// BcryptPasswordHashProxy checks the owner's password against its bcrypt hash.
type BcryptPasswordHashProxy struct{}

func NewBcryptPasswordHashProxy() *BcryptPasswordHashProxy {
	return &BcryptPasswordHashProxy{}
}

// Matches spends the full bcrypt cost against a fixed dummy hash when no hash is configured, so a sign-in
// against an unconfigured account takes as long as any other refusal.
func (hashProxy *BcryptPasswordHashProxy) Matches(passwordHash string, password string) bool {
	if passwordHash == "" {
		_ = bcrypt.CompareHashAndPassword(unconfiguredPasswordHash, []byte(password))
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}

// unconfiguredPasswordHash is the hash of 32 random bytes nobody kept: comparing against it always fails, at full cost.
var unconfiguredPasswordHash = []byte("$2a$10$Fduy4dHHzQFYyUjnlH067uHA7fC4m5tKYdXYYAlQ6kFauynoDJgmK")
