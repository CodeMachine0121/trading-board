package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApiKeyRepository struct {
	database ambientTransactionDatabase
}

func NewApiKeyRepository(database *gorm.DB) *ApiKeyRepository {
	return &ApiKeyRepository{database: ambientTransactionDatabase{root: database}}
}

func (keyRepository *ApiKeyRepository) Create(
	executionContext context.Context, apiKey entities.ApiKey,
) (entities.ApiKey, error) {
	if createError := keyRepository.database.within(executionContext).Create(&apiKey).Error; createError != nil {
		return entities.ApiKey{}, fmt.Errorf("create api key for %s: %w", apiKey.MemberName, createError)
	}

	return apiKey, nil
}

func (keyRepository *ApiKeyRepository) List(executionContext context.Context) ([]entities.ApiKey, error) {
	apiKeys := []entities.ApiKey{}
	if listError := keyRepository.database.within(executionContext).
		Order("created_at DESC, id DESC").Find(&apiKeys).Error; listError != nil {
		return nil, fmt.Errorf("list api keys: %w", listError)
	}

	return apiKeys, nil
}

// FindByID checks the identity's shape first, since PostgreSQL would answer a malformed one with an error, not "none".
func (keyRepository *ApiKeyRepository) FindByID(
	executionContext context.Context, apiKeyID string,
) (entities.ApiKey, bool, error) {
	if uuid.Validate(apiKeyID) != nil {
		return entities.ApiKey{}, false, nil
	}

	return keyRepository.found(keyRepository.database.within(executionContext).
		Where(&entities.ApiKey{ID: apiKeyID}))
}

func (keyRepository *ApiKeyRepository) FindByDigest(
	executionContext context.Context, keyDigest string,
) (entities.ApiKey, bool, error) {
	return keyRepository.found(keyRepository.database.within(executionContext).
		Where(&entities.ApiKey{KeyDigest: keyDigest}))
}

func (keyRepository *ApiKeyRepository) HasActiveMemberName(
	executionContext context.Context, memberNameKey string,
) (bool, error) {
	var activeCount int64
	if countError := keyRepository.database.within(executionContext).Model(&entities.ApiKey{}).
		Where(&entities.ApiKey{MemberNameKey: memberNameKey}).Where("revoked_at IS NULL").
		Count(&activeCount).Error; countError != nil {
		return false, fmt.Errorf("check member name in use: %w", countError)
	}

	return activeCount > 0, nil
}

// Revoke leaves a key already revoked as it is, so its revoke time never moves.
func (keyRepository *ApiKeyRepository) Revoke(
	executionContext context.Context, apiKeyID string, revokedAt time.Time,
) error {
	if updateError := keyRepository.database.within(executionContext).Model(&entities.ApiKey{}).
		Where(&entities.ApiKey{ID: apiKeyID}).Where("revoked_at IS NULL").
		Update("revoked_at", revokedAt).Error; updateError != nil {
		return fmt.Errorf("revoke api key %s: %w", apiKeyID, updateError)
	}

	return nil
}

func (keyRepository *ApiKeyRepository) TouchLastUsed(
	executionContext context.Context, apiKeyID string, usedAt time.Time,
) error {
	if updateError := keyRepository.database.within(executionContext).
		Model(&entities.ApiKey{ID: apiKeyID}).Update("last_used_at", usedAt).Error; updateError != nil {
		return fmt.Errorf("touch api key %s: %w", apiKeyID, updateError)
	}

	return nil
}

// found turns GORM's not-found error into the (false, nil) answer the interface promises.
func (keyRepository *ApiKeyRepository) found(query *gorm.DB) (entities.ApiKey, bool, error) {
	apiKey := entities.ApiKey{}
	findError := query.Take(&apiKey).Error
	if errors.Is(findError, gorm.ErrRecordNotFound) {
		return entities.ApiKey{}, false, nil
	}
	if findError != nil {
		return entities.ApiKey{}, false, fmt.Errorf("find api key: %w", findError)
	}

	return apiKey, true, nil
}
