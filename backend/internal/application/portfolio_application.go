package application

import (
	"context"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/service"
)

type PortfolioApplication struct {
	portfolioService *service.PortfolioService
}

func NewPortfolioApplication(portfolioService *service.PortfolioService) *PortfolioApplication {
	return &PortfolioApplication{portfolioService: portfolioService}
}

func (portfolioApplication *PortfolioApplication) ListPositions(
	executionContext context.Context, queryDto dto.PositionQueryDto,
) ([]dto.PositionDto, error) {
	return portfolioApplication.portfolioService.ListPositions(executionContext, queryDto)
}

func (portfolioApplication *PortfolioApplication) GetMonthSummary(
	executionContext context.Context, month string,
) (dto.MonthSummaryDto, error) {
	return portfolioApplication.portfolioService.GetMonthSummary(executionContext, month)
}
