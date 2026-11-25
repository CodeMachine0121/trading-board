package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderFillDto struct {
	ID       string          `json:"id"`
	Price    decimal.Decimal `json:"price"`
	Shares   int             `json:"shares"`
	Fee      decimal.Decimal `json:"fee"`
	Amount   decimal.Decimal `json:"amount"`
	FilledAt time.Time       `json:"filledAt"`
	// ReportedBy is empty for fills reported before attribution existed.
	ReportedBy string `json:"reportedBy"`
}
