package domains

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

const invalidCodeDescription = "授權碼無效或已過期"

// AuthorizationCodeDomain redeems a one-time code. The caller marks it consumed before asking, so a failed attempt
// uses the code up as well.
type AuthorizationCodeDomain struct {
	code entities.AuthorizationCode
}

func NewAuthorizationCodeDomain(code entities.AuthorizationCode) AuthorizationCodeDomain {
	return AuthorizationCodeDomain{code: code}
}

func (authorizationCode AuthorizationCodeDomain) IsConsumed() bool {
	return authorizationCode.code.ConsumedAt != nil
}

// Redeem admits the exchange only before expiry, for the same client and return address, and when the verifier
// hashes to the challenge (PKCE S256). Every refusal reads the same, so it tells an attacker nothing.
func (authorizationCode AuthorizationCodeDomain) Redeem(exchange dto.CodeExchangeDto, now time.Time) error {
	verifierDigest := sha256.Sum256([]byte(exchange.CodeVerifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(verifierDigest[:])

	if !now.Before(authorizationCode.code.ExpiresAt) ||
		exchange.ClientID != authorizationCode.code.ClientID ||
		exchange.RedirectUri != authorizationCode.code.RedirectUri ||
		exchange.CodeVerifier == "" ||
		subtle.ConstantTimeCompare([]byte(expectedChallenge), []byte(authorizationCode.code.CodeChallenge)) != 1 {
		return InvalidGrant(invalidCodeDescription)
	}

	return nil
}
