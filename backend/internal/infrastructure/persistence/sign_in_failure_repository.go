package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"gorm.io/gorm"
)

type SignInFailureRepository struct {
	database ambientTransactionDatabase
}

func NewSignInFailureRepository(database *gorm.DB) *SignInFailureRepository {
	return &SignInFailureRepository{database: ambientTransactionDatabase{root: database}}
}

func (failureRepository *SignInFailureRepository) Record(executionContext context.Context, failedAt time.Time) error {
	if createError := failureRepository.database.within(executionContext).
		Create(&entities.SignInFailure{FailedAt: failedAt}).Error; createError != nil {
		return fmt.Errorf("record sign-in failure: %w", createError)
	}

	return nil
}

func (failureRepository *SignInFailureRepository) ListLatest(
	executionContext context.Context, limit int,
) ([]entities.SignInFailure, error) {
	failures := []entities.SignInFailure{}
	if listError := failureRepository.database.within(executionContext).
		Order("failed_at DESC").Limit(limit).Find(&failures).Error; listError != nil {
		return nil, fmt.Errorf("list sign-in failures: %w", listError)
	}

	return failures, nil
}

// Clear forgets every failure; GORM refuses an unconditioned delete, hence the always-true condition.
func (failureRepository *SignInFailureRepository) Clear(executionContext context.Context) error {
	if deleteError := failureRepository.database.within(executionContext).
		Where("1 = 1").Delete(&entities.SignInFailure{}).Error; deleteError != nil {
		return fmt.Errorf("clear sign-in failures: %w", deleteError)
	}

	return nil
}
