package service

import (
	"context"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
)

// PortfolioService answers the figures derived from fills: holdings and the monthly summary.
type PortfolioService struct {
	ledgerLoader *TradeLedgerLoader
}

func NewPortfolioService(ledgerLoader *TradeLedgerLoader) *PortfolioService {
	return &PortfolioService{ledgerLoader: ledgerLoader}
}

func (portfolioService *PortfolioService) ListPositions(
	executionContext context.Context, queryDto dto.PositionQueryDto,
) ([]dto.PositionDto, error) {
	ledger, loadError := portfolioService.ledgerLoader.Load(executionContext)
	if loadError != nil {
		return nil, loadError
	}

	return ledger.Positions(queryDto)
}

func (portfolioService *PortfolioService) GetMonthSummary(
	executionContext context.Context, month string,
) (dto.MonthSummaryDto, error) {
	ledger, loadError := portfolioService.ledgerLoader.Load(executionContext)
	if loadError != nil {
		return dto.MonthSummaryDto{}, loadError
	}

	return ledger.MonthSummary(month)
}
