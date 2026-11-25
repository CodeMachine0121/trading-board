package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"gorm.io/gorm"
)

type TradeCardRepository struct {
	database ambientTransactionDatabase
}

func NewTradeCardRepository(database *gorm.DB) *TradeCardRepository {
	return &TradeCardRepository{database: ambientTransactionDatabase{root: database}}
}

// withOrdersAndFills loads a card's orders in published order and each order's fills in reported order.
func (cardRepository *TradeCardRepository) withOrdersAndFills(executionContext context.Context) *gorm.DB {
	return cardRepository.database.within(executionContext).
		Preload("Orders", func(orders *gorm.DB) *gorm.DB { return orders.Order("sequence ASC") }).
		Preload("Orders.Fills", func(fills *gorm.DB) *gorm.DB { return fills.Order("recorded_at ASC, id ASC") })
}

func (cardRepository *TradeCardRepository) FindByTradeDate(
	executionContext context.Context, tradeDate string,
) (entities.TradeCard, bool, error) {
	card := entities.TradeCard{}
	findError := cardRepository.withOrdersAndFills(executionContext).
		Where(&entities.TradeCard{TradeDate: tradeDate}).Take(&card).Error

	return cardRepository.found(card, findError)
}

func (cardRepository *TradeCardRepository) FindLatest(executionContext context.Context) (entities.TradeCard, bool, error) {
	card := entities.TradeCard{}
	findError := cardRepository.withOrdersAndFills(executionContext).Order("trade_date DESC").Take(&card).Error

	return cardRepository.found(card, findError)
}

// ListByMonth compares session dates as text, which orders correctly because they are all YYYY-MM-DD.
func (cardRepository *TradeCardRepository) ListByMonth(
	executionContext context.Context, month string,
) ([]entities.TradeCard, error) {
	firstDay, parseError := time.Parse("2006-01", month)
	if parseError != nil {
		return nil, fmt.Errorf("read month %q: %w", month, parseError)
	}

	cards := []entities.TradeCard{}
	listError := cardRepository.withOrdersAndFills(executionContext).
		Where("trade_date >= ? AND trade_date < ?",
			firstDay.Format("2006-01-02"), firstDay.AddDate(0, 1, 0).Format("2006-01-02")).
		Order("trade_date DESC").Find(&cards).Error
	if listError != nil {
		return nil, fmt.Errorf("list trade cards of %s: %w", month, listError)
	}

	return cards, nil
}

func (cardRepository *TradeCardRepository) FindEarliestTradeDate(executionContext context.Context) (string, bool, error) {
	card := entities.TradeCard{}
	findError := cardRepository.database.within(executionContext).
		Select("trade_date").Order("trade_date ASC").Take(&card).Error
	found, isFound, foundError := cardRepository.found(card, findError)

	return found.TradeDate, isFound, foundError
}

func (cardRepository *TradeCardRepository) Create(
	executionContext context.Context, card entities.TradeCard,
) (entities.TradeCard, error) {
	if createError := cardRepository.database.within(executionContext).Create(&card).Error; createError != nil {
		return entities.TradeCard{}, fmt.Errorf("create trade card %s: %w", card.TradeDate, createError)
	}

	return card, nil
}

func (cardRepository *TradeCardRepository) Delete(executionContext context.Context, cardID string) error {
	if deleteError := cardRepository.database.within(executionContext).
		Delete(&entities.TradeCard{ID: cardID}).Error; deleteError != nil {
		return fmt.Errorf("delete trade card %s: %w", cardID, deleteError)
	}

	return nil
}

// found turns GORM's not-found error into the (false, nil) answer the interface promises.
func (cardRepository *TradeCardRepository) found(
	card entities.TradeCard, findError error,
) (entities.TradeCard, bool, error) {
	if errors.Is(findError, gorm.ErrRecordNotFound) {
		return entities.TradeCard{}, false, nil
	}
	if findError != nil {
		return entities.TradeCard{}, false, fmt.Errorf("find trade card: %w", findError)
	}

	return card, true, nil
}
