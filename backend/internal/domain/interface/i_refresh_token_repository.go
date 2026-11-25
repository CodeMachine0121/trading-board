package _interface

import (
	"context"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

//go:generate go tool mockgen -source=i_refresh_token_repository.go -destination=mocks/mock_i_refresh_token_repository.go -package=mocks

// IRefreshTokenRepository stores the rotation chains of refresh tokens by digest.
type IRefreshTokenRepository interface {
	Create(executionContext context.Context, token entities.RefreshToken) error
	FindByDigest(executionContext context.Context, tokenDigest string) (entities.RefreshToken, bool, error)
	MarkRotated(executionContext context.Context, tokenID string, rotatedAt time.Time) error
	// RevokeFamily revokes every token of one sign-in that is not revoked yet.
	RevokeFamily(executionContext context.Context, familyID string, revokedAt time.Time) error
}
