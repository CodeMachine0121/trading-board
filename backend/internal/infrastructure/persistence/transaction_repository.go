package persistence

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// ledgerLockKey names the one advisory lock every write use case takes; any constant unique to this service works.
const ledgerLockKey int64 = 0x7472616465 // "trade"

type TransactionRepository struct {
	database *gorm.DB
}

func NewTransactionRepository(database *gorm.DB) *TransactionRepository {
	return &TransactionRepository{database: database}
}

// Exclusively holds a transaction-scoped advisory lock for the whole of work, so writes run one at a time and
// the lock is let go with the commit or rollback. It is the one hand-written statement outside the migrator:
// GORM has no API for advisory locks. The key is passed as a parameter, never spliced in.
func (transactionRepository *TransactionRepository) Exclusively(
	executionContext context.Context, work func(transactionContext context.Context) error,
) error {
	if _, isInsideTransaction := executionContext.Value(transactionContextKey{}).(*gorm.DB); isInsideTransaction {
		return work(executionContext)
	}

	return transactionRepository.database.WithContext(executionContext).Transaction(func(transaction *gorm.DB) error {
		if lockError := transaction.Exec("SELECT pg_advisory_xact_lock(?)", ledgerLockKey).Error; lockError != nil {
			return fmt.Errorf("wait for other writes: %w", lockError)
		}

		return work(context.WithValue(executionContext, transactionContextKey{}, transaction))
	})
}
