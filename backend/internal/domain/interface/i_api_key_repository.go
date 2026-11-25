package _interface

import (
	"context"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

//go:generate go tool mockgen -source=i_api_key_repository.go -destination=mocks/mock_i_api_key_repository.go -package=mocks

// IApiKeyRepository stores agent-team members' keys.
type IApiKeyRepository interface {
	// Create returns the key with its identity assigned.
	Create(executionContext context.Context, apiKey entities.ApiKey) (entities.ApiKey, error)
	// List returns every key, active and revoked, newest first.
	List(executionContext context.Context) ([]entities.ApiKey, error)
	// FindByID answers not-found (false) for an identity that is not even well formed.
	FindByID(executionContext context.Context, apiKeyID string) (entities.ApiKey, bool, error)
	FindByDigest(executionContext context.Context, keyDigest string) (entities.ApiKey, bool, error)
	HasActiveMemberName(executionContext context.Context, memberNameKey string) (bool, error)
	Revoke(executionContext context.Context, apiKeyID string, revokedAt time.Time) error
	TouchLastUsed(executionContext context.Context, apiKeyID string, usedAt time.Time) error
}
