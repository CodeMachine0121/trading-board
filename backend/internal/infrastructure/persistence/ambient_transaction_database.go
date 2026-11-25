package persistence

import (
	"context"

	"gorm.io/gorm"
)

// transactionContextKey carries the transaction TransactionRepository opened down to the repositories taking part.
type transactionContextKey struct{}

// ambientTransactionDatabase lets a repository join whatever transaction its caller opened, and work alone otherwise.
type ambientTransactionDatabase struct {
	root *gorm.DB
}

func (ambientDatabase ambientTransactionDatabase) within(executionContext context.Context) *gorm.DB {
	transaction, isInsideTransaction := executionContext.Value(transactionContextKey{}).(*gorm.DB)
	if isInsideTransaction {
		return transaction.WithContext(executionContext)
	}

	return ambientDatabase.root.WithContext(executionContext)
}
