package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderFillReportDto is what the owner reports; an invalid Fee means "estimate it for me".
type OrderFillReportDto struct {
	Price    decimal.Decimal
	Shares   int
	Fee      decimal.NullDecimal
	FilledAt time.Time
}
