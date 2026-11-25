package domains

import (
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

// RefreshTokenDomain says what a presented refresh token is: the current link of its chain, one already replaced
// (a reuse, which means theft), or one that no longer admits anybody.
type RefreshTokenDomain struct {
	token entities.RefreshToken
}

func NewRefreshTokenDomain(token entities.RefreshToken) RefreshTokenDomain {
	return RefreshTokenDomain{token: token}
}

func (refreshToken RefreshTokenDomain) FamilyID() string {
	return refreshToken.token.FamilyID
}

// IsReplaced is true for a token that was already rotated: presenting it again is treated as theft.
func (refreshToken RefreshTokenDomain) IsReplaced() bool {
	return refreshToken.token.RotatedAt != nil
}

// AdmitsAt lets the token renew only while it is neither revoked, nor replaced, nor 30 days old.
func (refreshToken RefreshTokenDomain) AdmitsAt(now time.Time) bool {
	return refreshToken.token.RevokedAt == nil && refreshToken.token.RotatedAt == nil &&
		now.Before(refreshToken.token.ExpiresAt)
}
