package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"gorm.io/gorm"
)

type AuthorizationCodeRepository struct {
	database ambientTransactionDatabase
}

func NewAuthorizationCodeRepository(database *gorm.DB) *AuthorizationCodeRepository {
	return &AuthorizationCodeRepository{database: ambientTransactionDatabase{root: database}}
}

func (codeRepository *AuthorizationCodeRepository) Create(
	executionContext context.Context, code entities.AuthorizationCode,
) error {
	if createError := codeRepository.database.within(executionContext).Create(&code).Error; createError != nil {
		return fmt.Errorf("create authorization code: %w", createError)
	}

	return nil
}

func (codeRepository *AuthorizationCodeRepository) FindByDigest(
	executionContext context.Context, codeDigest string,
) (entities.AuthorizationCode, bool, error) {
	code := entities.AuthorizationCode{}
	findError := codeRepository.database.within(executionContext).
		Where(&entities.AuthorizationCode{CodeDigest: codeDigest}).Take(&code).Error
	if errors.Is(findError, gorm.ErrRecordNotFound) {
		return entities.AuthorizationCode{}, false, nil
	}
	if findError != nil {
		return entities.AuthorizationCode{}, false, fmt.Errorf("find authorization code: %w", findError)
	}

	return code, true, nil
}

func (codeRepository *AuthorizationCodeRepository) MarkConsumed(
	executionContext context.Context, codeID string, consumedAt time.Time,
) error {
	if updateError := codeRepository.database.within(executionContext).
		Model(&entities.AuthorizationCode{ID: codeID}).Update("consumed_at", consumedAt).Error; updateError != nil {
		return fmt.Errorf("mark authorization code consumed: %w", updateError)
	}

	return nil
}
