package _interface

import (
	"context"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

//go:generate go tool mockgen -source=i_sign_in_failure_repository.go -destination=mocks/mock_i_sign_in_failure_repository.go -package=mocks

// ISignInFailureRepository remembers refused sign-ins until a successful one clears them.
type ISignInFailureRepository interface {
	Record(executionContext context.Context, failedAt time.Time) error
	// ListLatest returns at most limit failures, newest first.
	ListLatest(executionContext context.Context, limit int) ([]entities.SignInFailure, error)
	Clear(executionContext context.Context) error
}
