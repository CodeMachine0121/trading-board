package models

import (
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/shopspring/decimal"
)

// BuyOrderRequest takes decimals as strings ("52.5"); a bare number is accepted too.
type BuyOrderRequest struct {
	Symbol               string          `json:"symbol"`
	Name                 string          `json:"name"`
	LimitPrice           decimal.Decimal `json:"limitPrice"`
	MaxChasePrice        decimal.Decimal `json:"maxChasePrice"`
	Shares               int             `json:"shares"`
	EstimatedAmount      decimal.Decimal `json:"estimatedAmount"`
	StopLossPrice        decimal.Decimal `json:"stopLossPrice"`
	TakeProfitPrice      decimal.Decimal `json:"takeProfitPrice"`
	Reason               string          `json:"reason"`
	IsMonthEndAllocation bool            `json:"isMonthEndAllocation"`
}

func (buyOrderRequest BuyOrderRequest) ToWriteDto() dto.BuyOrderWriteDto {
	return dto.BuyOrderWriteDto{
		Symbol:               buyOrderRequest.Symbol,
		Name:                 buyOrderRequest.Name,
		LimitPrice:           buyOrderRequest.LimitPrice,
		MaxChasePrice:        buyOrderRequest.MaxChasePrice,
		Shares:               buyOrderRequest.Shares,
		EstimatedAmount:      buyOrderRequest.EstimatedAmount,
		StopLossPrice:        buyOrderRequest.StopLossPrice,
		TakeProfitPrice:      buyOrderRequest.TakeProfitPrice,
		Reason:               buyOrderRequest.Reason,
		IsMonthEndAllocation: buyOrderRequest.IsMonthEndAllocation,
	}
}
