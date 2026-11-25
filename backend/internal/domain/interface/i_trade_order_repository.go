package _interface

import (
	"context"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

//go:generate go tool mockgen -source=i_trade_order_repository.go -destination=mocks/mock_i_trade_order_repository.go -package=mocks

// ITradeOrderRepository reads orders with their fills and records who skipped an order.
type ITradeOrderRepository interface {
	// FindByID answers not-found (false) for an identity that is not even well formed.
	FindByID(executionContext context.Context, orderID string) (entities.TradeOrder, bool, error)
	// ListWithFills returns every order that has at least one fill, fills included: the whole fill log.
	ListWithFills(executionContext context.Context) ([]entities.TradeOrder, error)
	// MarkSkipped leaves an order that is already skipped as it is, including who skipped it.
	MarkSkipped(executionContext context.Context, orderID string, skippedBy string) error
}
