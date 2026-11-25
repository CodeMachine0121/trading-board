package dto

import "github.com/shopspring/decimal"

// OrderDto serves both sides; side-only fields are pointers so a buy never shows an exit reason
// and a sell never shows a stop-loss, while a buy's false month-end flag still appears.
type OrderDto struct {
	ID                   string           `json:"id"`
	Side                 string           `json:"side"`
	Symbol               string           `json:"symbol"`
	Name                 string           `json:"name"`
	LimitPrice           decimal.Decimal  `json:"limitPrice"`
	MaxChasePrice        *decimal.Decimal `json:"maxChasePrice,omitempty"`
	Shares               int              `json:"shares"`
	EstimatedAmount      *decimal.Decimal `json:"estimatedAmount,omitempty"`
	StopLossPrice        *decimal.Decimal `json:"stopLossPrice,omitempty"`
	TakeProfitPrice      *decimal.Decimal `json:"takeProfitPrice,omitempty"`
	ExitReason           *string          `json:"exitReason,omitempty"`
	Reason               string           `json:"reason"`
	IsMonthEndAllocation *bool            `json:"isMonthEndAllocation,omitempty"`
	Status               string           `json:"status"`
	FilledShares         int              `json:"filledShares"`
	Fills                []OrderFillDto   `json:"fills"`
	// SkippedBy is empty unless the order was skipped (and for skips recorded before attribution existed).
	SkippedBy string `json:"skippedBy"`
}
