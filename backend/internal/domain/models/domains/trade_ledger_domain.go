package domains

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/shopspring/decimal"
)

const (
	defaultMonthEndStartDay = 22
	percentDecimalPlaces    = 2
)

var oneHundred = decimal.NewFromInt(100)

// TradeLedgerDomain replays every reported fill, in the order reported, into holdings and monthly books.
// Nothing it answers is stored anywhere, so no figure can drift from the fills it came from.
type TradeLedgerDomain struct {
	entries          []vo.LedgerEntryVo
	earliestCardDate string
	calendar         BusinessCalendarDomain
	policy           vo.BudgetPolicyVo
	now              time.Time

	positions     []*PositionDomain
	openPositions map[string]*PositionDomain
	monthlyBooks  map[string]*MonthlyBookDomain
}

// NewTradeLedgerDomain takes the earliest card's session date ("" when there is none), since a card alone
// already makes its month the first month. A month-end start day outside 1..31 falls back to the 22nd.
func NewTradeLedgerDomain(
	entries []vo.LedgerEntryVo, earliestCardDate string, calendar BusinessCalendarDomain,
	policy vo.BudgetPolicyVo, now time.Time,
) TradeLedgerDomain {
	if policy.MonthEndStartDay < 1 || policy.MonthEndStartDay > 31 {
		policy.MonthEndStartDay = defaultMonthEndStartDay
	}
	policy.MonthlyAllocation = decimal.Max(policy.MonthlyAllocation, decimal.Zero)

	sortedEntries := slices.Clone(entries)
	slices.SortStableFunc(sortedEntries, func(left vo.LedgerEntryVo, right vo.LedgerEntryVo) int {
		return cmp.Or(left.RecordedAt.Compare(right.RecordedAt), cmp.Compare(left.FillID, right.FillID))
	})

	ledger := TradeLedgerDomain{
		entries:          sortedEntries,
		earliestCardDate: earliestCardDate,
		calendar:         calendar,
		policy:           policy,
		now:              now,
		openPositions:    map[string]*PositionDomain{},
		monthlyBooks:     map[string]*MonthlyBookDomain{},
	}

	for _, entry := range sortedEntries {
		month := calendar.MonthOf(entry.FilledAt)
		monthlyBook, hasBook := ledger.monthlyBooks[month]
		if !hasBook {
			monthlyBook = &MonthlyBookDomain{}
			ledger.monthlyBooks[month] = monthlyBook
		}

		position, isHeld := ledger.openPositions[entry.Symbol]
		if entry.Side == vo.OrderSideBuy {
			if !isHeld {
				position = NewPositionDomain(entry.Symbol, calendar.DateOf(entry.FilledAt))
				ledger.positions = append(ledger.positions, position)
				ledger.openPositions[entry.Symbol] = position
			}
			position.Buy(entry)
			monthlyBook.RecordPurchase(entry)
			continue
		}

		// A stored sale always had a holding to come from, because Admit refuses any other.
		if !isHeld {
			continue
		}
		realized, costOfSold := position.Sell(entry, calendar.DateOf(entry.FilledAt))
		monthlyBook.RecordSale(entry, realized, costOfSold)
		if position.IsClosed() {
			delete(ledger.openPositions, entry.Symbol)
		}
	}

	return ledger
}

func (ledger TradeLedgerDomain) HeldShares(symbol string) int {
	position, isHeld := ledger.openPositions[symbol]
	if !isHeld {
		return 0
	}

	return position.Shares()
}

// RequireHoldings refuses sales that, added up per stock, exceed what is held now.
func (ledger TradeLedgerDomain) RequireHoldings(soldSharesBySymbol map[string]int) error {
	symbols := make([]string, 0, len(soldSharesBySymbol))
	for symbol := range soldSharesBySymbol {
		symbols = append(symbols, symbol)
	}
	slices.Sort(symbols)

	for _, symbol := range symbols {
		heldShares := ledger.HeldShares(symbol)
		if heldShares == 0 {
			return ValidationFailed("sellExceedsHolding", fmt.Sprintf("目前沒有持有 %s，不能賣出", symbol))
		}
		if soldSharesBySymbol[symbol] > heldShares {
			return ValidationFailed("sellExceedsHolding", fmt.Sprintf(
				"%s 目前持有 %d 股，賣出委託合計 %d 股超過持股", symbol, heldShares, soldSharesBySymbol[symbol]))
		}
	}

	return nil
}

// Admit decides whether a new fill may join the ledger: a sale may not exceed the holding, and a purchase
// may not leave its month, or any later month the money carries into, with a negative remaining.
func (ledger TradeLedgerDomain) Admit(entry vo.LedgerEntryVo) error {
	if entry.Side == vo.OrderSideSell {
		heldShares := ledger.HeldShares(entry.Symbol)
		if entry.Shares > heldShares {
			return Conflict("insufficientShares", fmt.Sprintf(
				"%s 目前只持有 %d 股，不能賣出 %d 股", entry.Symbol, heldShares, entry.Shares))
		}

		return nil
	}

	withEntry := NewTradeLedgerDomain(
		append(slices.Clone(ledger.entries), entry), ledger.earliestCardDate, ledger.calendar, ledger.policy, ledger.now)
	fillMonth := ledger.calendar.MonthOf(entry.FilledAt)
	// The walk reaches the current month, or any later month a fill was dated into.
	lastMonth := max(ledger.calendar.MonthOf(ledger.now), fillMonth)
	for month := range ledger.monthlyBooks {
		lastMonth = max(lastMonth, month)
	}

	var shortfall error
	withEntry.walkBudgets(lastMonth, func(month string, _ decimal.Decimal, remaining decimal.Decimal) {
		if shortfall == nil && month >= fillMonth && remaining.IsNegative() {
			shortfall = Conflict("insufficientBudget", fmt.Sprintf(
				"%s 的預算不足：這筆買進會讓剩餘金額變成 %s", month, remaining.String()))
		}
	})

	return shortfall
}

// Positions lists open holdings in the order they opened, or closed holdings newest close first.
func (ledger TradeLedgerDomain) Positions(query dto.PositionQueryDto) ([]dto.PositionDto, error) {
	status := vo.PositionStatusVo(query.Status)
	if status == "" {
		status = vo.PositionStatusOpen
	}
	if status != vo.PositionStatusOpen && status != vo.PositionStatusClosed {
		return nil, ValidationFailed("invalidRequest", "status 只能是 open 或 closed")
	}
	if query.Month != "" && !ledger.calendar.IsMonth(query.Month) {
		return nil, ValidationFailed("invalidRequest", "month 須為 YYYY-MM")
	}

	today := ledger.calendar.DateOf(ledger.now)
	positionDtos := []dto.PositionDto{}
	if status == vo.PositionStatusOpen {
		for _, position := range ledger.positions {
			if !position.IsClosed() {
				positionDtos = append(positionDtos, position.ToDto(ledger.calendar, today))
			}
		}

		return positionDtos, nil
	}

	closedPositions := []*PositionDomain{}
	for _, position := range ledger.positions {
		if position.IsClosed() && (query.Month == "" || position.ClosedDate()[:7] == query.Month) {
			closedPositions = append(closedPositions, position)
		}
	}
	slices.SortStableFunc(closedPositions, func(left *PositionDomain, right *PositionDomain) int {
		return cmp.Compare(right.ClosedDate(), left.ClosedDate())
	})
	for _, position := range closedPositions {
		positionDtos = append(positionDtos, position.ToDto(ledger.calendar, today))
	}

	return positionDtos, nil
}

// MonthSummary is the month's budget and performance; months before the first one do not exist.
func (ledger TradeLedgerDomain) MonthSummary(month string) (dto.MonthSummaryDto, error) {
	if !ledger.calendar.IsMonth(month) {
		return dto.MonthSummaryDto{}, ValidationFailed("invalidRequest", "月份須為 YYYY-MM")
	}
	if month < ledger.firstMonth() {
		return dto.MonthSummaryDto{}, NotFound("monthNotFound", fmt.Sprintf("%s 早於第一個月份，沒有資料", month))
	}

	summaryDto := dto.MonthSummaryDto{Month: month, Allocation: ledger.policy.MonthlyAllocation}
	ledger.walkBudgets(month, func(walkedMonth string, carriedOver decimal.Decimal, remaining decimal.Decimal) {
		summaryDto.CarriedOver = carriedOver
		summaryDto.Remaining = remaining
	})

	monthlyBook := ledger.bookOf(month)
	summaryDto.RecycledProceeds = monthlyBook.recycledProceeds
	summaryDto.Spent = monthlyBook.spent
	summaryDto.RealizedPnl = monthlyBook.realizedPnl

	currentMonth := ledger.calendar.MonthOf(ledger.now)
	switch {
	case month < currentMonth:
		summaryDto.Mode = string(vo.BudgetModeMonthEnd)
	case month > currentMonth:
		summaryDto.Mode = string(vo.BudgetModeStrict)
	case ledger.calendar.DayOfMonth(ledger.now) >= ledger.policy.MonthEndStartDay:
		summaryDto.Mode = string(vo.BudgetModeMonthEnd)
	default:
		summaryDto.Mode = string(vo.BudgetModeStrict)
	}

	for _, position := range ledger.positions {
		if !position.IsClosed() || position.ClosedDate()[:7] != month {
			continue
		}
		summaryDto.ClosedPositionCount++
		if position.Result() == vo.PositionResultWin {
			summaryDto.WinCount++
		} else {
			summaryDto.LossCount++
		}
	}

	if summaryDto.ClosedPositionCount > 0 {
		summaryDto.WinRate = decimal.NewFromInt(int64(summaryDto.WinCount)).Mul(oneHundred).
			Div(decimal.NewFromInt(int64(summaryDto.ClosedPositionCount))).Round(percentDecimalPlaces)
	}
	if monthlyBook.costOfSoldShares.IsPositive() {
		summaryDto.RealizedReturnPct = monthlyBook.realizedPnl.Mul(oneHundred).
			Div(monthlyBook.costOfSoldShares).Round(percentDecimalPlaces)
	}

	return summaryDto, nil
}

// walkBudgets visits every month from the first through lastMonth with what it carried in and what it has left.
func (ledger TradeLedgerDomain) walkBudgets(
	lastMonth string, visit func(month string, carriedOver decimal.Decimal, remaining decimal.Decimal),
) {
	carriedOver := decimal.Zero
	for month := ledger.firstMonth(); month <= lastMonth; month = ledger.calendar.NextMonth(month) {
		remaining := ledger.bookOf(month).Remaining(ledger.policy.MonthlyAllocation, carriedOver)
		visit(month, carriedOver, remaining)
		carriedOver = remaining
	}
}

// bookOf answers an empty book for a month without fills.
func (ledger TradeLedgerDomain) bookOf(month string) *MonthlyBookDomain {
	monthlyBook, hasBook := ledger.monthlyBooks[month]
	if !hasBook {
		return &MonthlyBookDomain{}
	}

	return monthlyBook
}

// firstMonth is the month of the earliest card or fill, or the current month when nothing exists yet.
func (ledger TradeLedgerDomain) firstMonth() string {
	months := make([]string, 0, len(ledger.monthlyBooks)+1)
	for month := range ledger.monthlyBooks {
		months = append(months, month)
	}
	if ledger.earliestCardDate != "" {
		months = append(months, ledger.earliestCardDate[:7])
	}
	if len(months) == 0 {
		return ledger.calendar.MonthOf(ledger.now)
	}

	return slices.Min(months)
}
