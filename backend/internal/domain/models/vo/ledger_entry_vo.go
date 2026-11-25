package vo

import (
	"time"

	"github.com/shopspring/decimal"
)

// LedgerEntryVo is one fill together with the order facts the ledger needs to replay it.
type LedgerEntryVo struct {
	FillID          string
	Side            OrderSideVo
	Symbol          string
	Name            string
	TradeDate       string
	StopLossPrice   decimal.NullDecimal
	TakeProfitPrice decimal.NullDecimal
	Shares          int
	Amount          decimal.Decimal
	Fee             decimal.Decimal
	FilledAt        time.Time
	RecordedAt      time.Time
}
