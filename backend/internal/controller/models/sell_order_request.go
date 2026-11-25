package models

import (
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/shopspring/decimal"
)

type SellOrderRequest struct {
	Symbol     string          `json:"symbol"`
	Name       string          `json:"name"`
	LimitPrice decimal.Decimal `json:"limitPrice"`
	Shares     int             `json:"shares"`
	ExitReason string          `json:"exitReason"`
	Reason     string          `json:"reason"`
}

func (sellOrderRequest SellOrderRequest) ToWriteDto() dto.SellOrderWriteDto {
	return dto.SellOrderWriteDto{
		Symbol:     sellOrderRequest.Symbol,
		Name:       sellOrderRequest.Name,
		LimitPrice: sellOrderRequest.LimitPrice,
		Shares:     sellOrderRequest.Shares,
		ExitReason: sellOrderRequest.ExitReason,
		Reason:     sellOrderRequest.Reason,
	}
}
