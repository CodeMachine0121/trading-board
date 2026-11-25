package persistence

import (
	"context"
	"fmt"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"gorm.io/gorm"
)

type OrderFillRepository struct {
	database ambientTransactionDatabase
}

func NewOrderFillRepository(database *gorm.DB) *OrderFillRepository {
	return &OrderFillRepository{database: ambientTransactionDatabase{root: database}}
}

func (fillRepository *OrderFillRepository) Create(
	executionContext context.Context, fill entities.OrderFill,
) (entities.OrderFill, error) {
	if createError := fillRepository.database.within(executionContext).Create(&fill).Error; createError != nil {
		return entities.OrderFill{}, fmt.Errorf("create fill of order %s: %w", fill.TradeOrderID, createError)
	}

	return fill, nil
}
