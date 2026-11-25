package _interface

import (
	"context"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

//go:generate go tool mockgen -source=i_authorization_code_repository.go -destination=mocks/mock_i_authorization_code_repository.go -package=mocks

// IAuthorizationCodeRepository stores one-time authorization codes by digest.
type IAuthorizationCodeRepository interface {
	Create(executionContext context.Context, code entities.AuthorizationCode) error
	FindByDigest(executionContext context.Context, codeDigest string) (entities.AuthorizationCode, bool, error)
	MarkConsumed(executionContext context.Context, codeID string, consumedAt time.Time) error
}
