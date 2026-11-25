package persistence

import (
	"context"
	"fmt"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"gorm.io/gorm"
)

// schemaMigrationLockKey keeps replicas that start together from migrating at the same time.
const schemaMigrationLockKey int64 = 0x6d696772617465 // "migrate"

// SchemaMigrator syncs the schema from the entities; the loose element type is what GORM's AutoMigrate takes.
type SchemaMigrator struct {
	database *gorm.DB
}

func NewSchemaMigrator(database *gorm.DB) *SchemaMigrator {
	return &SchemaMigrator{database: database}
}

// Migrate syncs every entity and returns the table names; register new entities in the slice below.
func (schemaMigrator *SchemaMigrator) Migrate() ([]string, error) {
	migratedEntities := []any{
		&entities.TradeCard{},
		&entities.TradeOrder{},
		&entities.OrderFill{},
	}

	if migrateError := schemaMigrator.database.AutoMigrate(migratedEntities...); migrateError != nil {
		return nil, fmt.Errorf("auto migrate schema: %w", migrateError)
	}

	return []string{
		entities.TradeCard{}.TableName(),
		entities.TradeOrder{}.TableName(),
		entities.OrderFill{}.TableName(),
	}, nil
}

// Exclusively runs work holding a session advisory lock on one connection; GORM has no API for advisory
// locks, so these two statements are written by hand, with the key passed as a parameter.
func (schemaMigrator *SchemaMigrator) Exclusively(executionContext context.Context, work func() error) error {
	return schemaMigrator.database.WithContext(executionContext).Connection(func(lockHolder *gorm.DB) error {
		if lockError := lockHolder.Exec("SELECT pg_advisory_lock(?)", schemaMigrationLockKey).Error; lockError != nil {
			return fmt.Errorf("wait for other migrations: %w", lockError)
		}
		defer lockHolder.Exec("SELECT pg_advisory_unlock(?)", schemaMigrationLockKey)

		return work()
	})
}
