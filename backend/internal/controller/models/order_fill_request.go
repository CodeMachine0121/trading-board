package models

import (
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/shopspring/decimal"
)

// OrderFillRequest leaves fee out (or null) to have it estimated.
type OrderFillRequest struct {
	Price    decimal.Decimal     `json:"price"`
	Shares   int                 `json:"shares"`
	Fee      decimal.NullDecimal `json:"fee"`
	FilledAt time.Time           `json:"filledAt"`
}

func (fillRequest OrderFillRequest) ToReportDto() dto.OrderFillReportDto {
	return dto.OrderFillReportDto{
		Price:    fillRequest.Price,
		Shares:   fillRequest.Shares,
		Fee:      fillRequest.Fee,
		FilledAt: fillRequest.FilledAt,
	}
}
