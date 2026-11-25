package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	database ambientTransactionDatabase
}

func NewRefreshTokenRepository(database *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{database: ambientTransactionDatabase{root: database}}
}

func (tokenRepository *RefreshTokenRepository) Create(
	executionContext context.Context, token entities.RefreshToken,
) error {
	if createError := tokenRepository.database.within(executionContext).Create(&token).Error; createError != nil {
		return fmt.Errorf("create refresh token: %w", createError)
	}

	return nil
}

func (tokenRepository *RefreshTokenRepository) FindByDigest(
	executionContext context.Context, tokenDigest string,
) (entities.RefreshToken, bool, error) {
	token := entities.RefreshToken{}
	findError := tokenRepository.database.within(executionContext).
		Where(&entities.RefreshToken{TokenDigest: tokenDigest}).Take(&token).Error
	if errors.Is(findError, gorm.ErrRecordNotFound) {
		return entities.RefreshToken{}, false, nil
	}
	if findError != nil {
		return entities.RefreshToken{}, false, fmt.Errorf("find refresh token: %w", findError)
	}

	return token, true, nil
}

func (tokenRepository *RefreshTokenRepository) MarkRotated(
	executionContext context.Context, tokenID string, rotatedAt time.Time,
) error {
	if updateError := tokenRepository.database.within(executionContext).
		Model(&entities.RefreshToken{ID: tokenID}).Update("rotated_at", rotatedAt).Error; updateError != nil {
		return fmt.Errorf("mark refresh token rotated: %w", updateError)
	}

	return nil
}

func (tokenRepository *RefreshTokenRepository) RevokeFamily(
	executionContext context.Context, familyID string, revokedAt time.Time,
) error {
	if updateError := tokenRepository.database.within(executionContext).
		Model(&entities.RefreshToken{}).
		Where(&entities.RefreshToken{FamilyID: familyID}).Where("revoked_at IS NULL").
		Update("revoked_at", revokedAt).Error; updateError != nil {
		return fmt.Errorf("revoke refresh token family: %w", updateError)
	}

	return nil
}
