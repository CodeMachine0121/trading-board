package _interface

import "context"

//go:generate go tool mockgen -source=i_transaction_repository.go -destination=mocks/mock_i_transaction_repository.go -package=mocks

// ITransactionRepository runs a write use case alone: one transaction, with every other write waiting,
// so rules spanning many rows (no oversell, no overspend, no replacing a filled card) cannot race.
type ITransactionRepository interface {
	// Exclusively commits when work returns nil and rolls back otherwise; inside another, it joins that one.
	Exclusively(executionContext context.Context, work func(transactionContext context.Context) error) error
}
