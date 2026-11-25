package domains

import (
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/shopspring/decimal"
)

// MonthlyBookDomain adds up one month's fills: money spent, money recycled, profit realized and the cost it released.
type MonthlyBookDomain struct {
	spent            decimal.Decimal
	recycledProceeds decimal.Decimal
	realizedPnl      decimal.Decimal
	costOfSoldShares decimal.Decimal
}

func (monthlyBook *MonthlyBookDomain) RecordPurchase(entry vo.LedgerEntryVo) {
	monthlyBook.spent = monthlyBook.spent.Add(entry.Amount).Add(entry.Fee)
}

func (monthlyBook *MonthlyBookDomain) RecordSale(entry vo.LedgerEntryVo, realized decimal.Decimal, costOfSold decimal.Decimal) {
	monthlyBook.recycledProceeds = monthlyBook.recycledProceeds.Add(entry.Amount).Sub(entry.Fee)
	monthlyBook.realizedPnl = monthlyBook.realizedPnl.Add(realized)
	monthlyBook.costOfSoldShares = monthlyBook.costOfSoldShares.Add(costOfSold)
}

// Remaining is what is left of the month given what came in from the month before.
func (monthlyBook *MonthlyBookDomain) Remaining(allocation decimal.Decimal, carriedOver decimal.Decimal) decimal.Decimal {
	return allocation.Add(carriedOver).Add(monthlyBook.recycledProceeds).Sub(monthlyBook.spent)
}
