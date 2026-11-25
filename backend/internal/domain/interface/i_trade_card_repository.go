package _interface

import (
	"context"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

//go:generate go tool mockgen -source=i_trade_card_repository.go -destination=mocks/mock_i_trade_card_repository.go -package=mocks

// ITradeCardRepository stores cards; every card it returns carries its orders and their fills.
type ITradeCardRepository interface {
	FindByTradeDate(executionContext context.Context, tradeDate string) (entities.TradeCard, bool, error)
	FindLatest(executionContext context.Context) (entities.TradeCard, bool, error)
	// ListByMonth returns the month's cards newest session date first.
	ListByMonth(executionContext context.Context, month string) ([]entities.TradeCard, error)
	FindEarliestTradeDate(executionContext context.Context) (string, bool, error)
	// Create stores the card with its orders and returns it with identities assigned.
	Create(executionContext context.Context, card entities.TradeCard) (entities.TradeCard, error)
	// Delete removes the card and its orders.
	Delete(executionContext context.Context, cardID string) error
}
