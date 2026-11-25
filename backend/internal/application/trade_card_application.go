package application

import (
	"context"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/service"
)

type TradeCardApplication struct {
	tradeCardService *service.TradeCardService
}

func NewTradeCardApplication(tradeCardService *service.TradeCardService) *TradeCardApplication {
	return &TradeCardApplication{tradeCardService: tradeCardService}
}

func (cardApplication *TradeCardApplication) PublishCard(
	executionContext context.Context, publishDto dto.TradeCardPublishDto,
) (dto.TradeCardDto, error) {
	return cardApplication.tradeCardService.PublishCard(executionContext, publishDto)
}

func (cardApplication *TradeCardApplication) GetCard(
	executionContext context.Context, tradeDate string,
) (dto.TradeCardDto, error) {
	return cardApplication.tradeCardService.GetCard(executionContext, tradeDate)
}

func (cardApplication *TradeCardApplication) GetLatestCard(executionContext context.Context) (dto.TradeCardDto, error) {
	return cardApplication.tradeCardService.GetLatestCard(executionContext)
}

func (cardApplication *TradeCardApplication) ListCards(
	executionContext context.Context, month string,
) ([]dto.TradeCardSummaryDto, error) {
	return cardApplication.tradeCardService.ListCards(executionContext, month)
}
