package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TradeOrderRepository struct {
	database ambientTransactionDatabase
}

func NewTradeOrderRepository(database *gorm.DB) *TradeOrderRepository {
	return &TradeOrderRepository{database: ambientTransactionDatabase{root: database}}
}

// FindByID checks the identity's shape first, since PostgreSQL would answer a malformed one with an error, not "none".
func (orderRepository *TradeOrderRepository) FindByID(
	executionContext context.Context, orderID string,
) (entities.TradeOrder, bool, error) {
	if uuid.Validate(orderID) != nil {
		return entities.TradeOrder{}, false, nil
	}

	order := entities.TradeOrder{}
	findError := orderRepository.database.within(executionContext).
		Preload("Fills", func(fills *gorm.DB) *gorm.DB { return fills.Order("recorded_at ASC, id ASC") }).
		Where(&entities.TradeOrder{ID: orderID}).Take(&order).Error
	if errors.Is(findError, gorm.ErrRecordNotFound) {
		return entities.TradeOrder{}, false, nil
	}
	if findError != nil {
		return entities.TradeOrder{}, false, fmt.Errorf("find trade order %s: %w", orderID, findError)
	}

	return order, true, nil
}

func (orderRepository *TradeOrderRepository) ListWithFills(executionContext context.Context) ([]entities.TradeOrder, error) {
	orders := []entities.TradeOrder{}
	listError := orderRepository.database.within(executionContext).
		Preload("Fills").
		Where("EXISTS (?)", orderRepository.database.within(executionContext).
			Model(&entities.OrderFill{}).Select("1").Where(`"OrderFills".trade_order_id = "TradeOrders".id`)).
		Find(&orders).Error
	if listError != nil {
		return nil, fmt.Errorf("list trade orders with fills: %w", listError)
	}

	return orders, nil
}

func (orderRepository *TradeOrderRepository) MarkSkipped(
	executionContext context.Context, orderID string, skippedBy string,
) error {
	if updateError := orderRepository.database.within(executionContext).Model(&entities.TradeOrder{}).
		Where(&entities.TradeOrder{ID: orderID}).Where("is_skipped = ?", false).
		Updates(map[string]any{"is_skipped": true, "skipped_by": skippedBy}).Error; updateError != nil {
		return fmt.Errorf("mark trade order %s skipped: %w", orderID, updateError)
	}

	return nil
}
