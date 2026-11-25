package domains

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

// displayPrefixLength keeps "tbk_" plus eight random characters: enough to tell keys apart, useless to an attacker.
const displayPrefixLength = 12

// OpaqueSecretDomain is a high-entropy secret handed out once (code, refresh token, API key). The board keeps only
// its SHA-256 digest: with 256 random bits a plain digest is as good as a slow hash and can be looked up directly.
type OpaqueSecretDomain struct {
	value string
}

func NewOpaqueSecretDomain(value string) OpaqueSecretDomain {
	return OpaqueSecretDomain{value: value}
}

func (secret OpaqueSecretDomain) Value() string {
	return secret.value
}

func (secret OpaqueSecretDomain) Digest() string {
	digest := sha256.Sum256([]byte(secret.value))

	return hex.EncodeToString(digest[:])
}

func (secret OpaqueSecretDomain) DisplayPrefix() string {
	if len(secret.value) <= displayPrefixLength {
		return secret.value
	}

	return secret.value[:displayPrefixLength]
}

// ToRefreshToken starts a new sign-in's chain when familyID is empty, and continues that chain otherwise.
func (secret OpaqueSecretDomain) ToRefreshToken(
	familyID string, issuedAt time.Time, lifetime time.Duration,
) entities.RefreshToken {
	if familyID == "" {
		familyID = secret.Digest()
	}

	return entities.RefreshToken{
		FamilyID:    familyID,
		TokenDigest: secret.Digest(),
		IssuedAt:    issuedAt,
		ExpiresAt:   issuedAt.Add(lifetime),
	}
}
