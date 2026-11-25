package domains_test

import (
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/shopspring/decimal"
)

var taipei, _ = time.LoadLocation("Asia/Taipei")

var calendar = domains.NewBusinessCalendarDomain(taipei)

var defaultBudgetPolicy = vo.BudgetPolicyVo{MonthlyAllocation: decimal.NewFromInt(10000), MonthEndStartDay: 22}

var defaultFeeSchedule = vo.FeeScheduleVo{
	BrokerFeeRate:      decimal.RequireFromString("0.001425"),
	Discount:           decimal.RequireFromString("0.6"),
	MinimumFee:         decimal.RequireFromString("1"),
	TransactionTaxRate: decimal.RequireFromString("0.003"),
}

func money(text string) decimal.Decimal {
	return decimal.RequireFromString(text)
}

// taipeiTime reads "2026-11-26 09:05" in Taipei time.
func taipeiTime(text string) time.Time {
	parsed, parseError := time.ParseInLocation("2006-01-02 15:04", text, taipei)
	if parseError != nil {
		panic(parseError)
	}

	return parsed
}

// ledgerBook builds entries whose report order is the order they are added in.
type ledgerBook struct {
	entries []vo.LedgerEntryVo
}

func (book *ledgerBook) add(side vo.OrderSideVo, symbol string, tradeDate string, shares int, price string, fee string, filledAt string) *ledgerBook {
	book.entries = append(book.entries, book.entryOf(side, symbol, tradeDate, shares, price, fee, filledAt))

	return book
}

func (book *ledgerBook) entryOf(side vo.OrderSideVo, symbol string, tradeDate string, shares int, price string, fee string, filledAt string) vo.LedgerEntryVo {
	return vo.LedgerEntryVo{
		FillID:          symbol + string(rune('a'+len(book.entries))),
		Side:            side,
		Symbol:          symbol,
		Name:            "name-" + symbol,
		TradeDate:       tradeDate,
		StopLossPrice:   decimal.NewNullDecimal(money("48")),
		TakeProfitPrice: decimal.NewNullDecimal(money("60")),
		Shares:          shares,
		Amount:          money(price).Mul(decimal.NewFromInt(int64(shares))),
		Fee:             money(fee),
		FilledAt:        taipeiTime(filledAt),
		RecordedAt:      time.Date(2026, 1, 1, 0, 0, len(book.entries), 0, time.UTC),
	}
}

func (book *ledgerBook) ledgerAt(now string) domains.TradeLedgerDomain {
	return domains.NewTradeLedgerDomain(book.entries, "", calendar, defaultBudgetPolicy, taipeiTime(now))
}
