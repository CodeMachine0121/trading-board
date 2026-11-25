package application

import (
	"context"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/service"
)

type TradeOrderApplication struct {
	tradeOrderService *service.TradeOrderService
}

func NewTradeOrderApplication(tradeOrderService *service.TradeOrderService) *TradeOrderApplication {
	return &TradeOrderApplication{tradeOrderService: tradeOrderService}
}

func (orderApplication *TradeOrderApplication) ReportFill(
	executionContext context.Context, orderID string, reportDto dto.OrderFillReportDto,
) (dto.OrderDto, error) {
	return orderApplication.tradeOrderService.ReportFill(executionContext, orderID, reportDto)
}

func (orderApplication *TradeOrderApplication) SkipOrder(
	executionContext context.Context, orderID string,
) (dto.OrderDto, error) {
	return orderApplication.tradeOrderService.SkipOrder(executionContext, orderID)
}
