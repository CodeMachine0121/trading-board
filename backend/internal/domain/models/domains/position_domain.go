package domains

import (
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/shopspring/decimal"
)

const (
	averageCostDecimalPlaces = 4
	costOfSoldDecimalPlaces  = 2
)

// PositionDomain is one holding of one stock, from the buy fill that opened it to the sell fill that emptied it.
type PositionDomain struct {
	symbol             string
	name               string
	shares             int
	totalCost          decimal.Decimal
	boughtShares       int
	totalBuyCost       decimal.Decimal
	realizedPnl        decimal.Decimal
	openedDate         string
	closedDate         string
	latestBuyTradeDate string
	stopLossPrice      decimal.NullDecimal
	takeProfitPrice    decimal.NullDecimal
}

func NewPositionDomain(symbol string, openedDate string) *PositionDomain {
	return &PositionDomain{symbol: symbol, openedDate: openedDate}
}

// Buy adds shares at their cost including the fee; stop-loss and take-profit follow the latest buy order by session date.
func (position *PositionDomain) Buy(entry vo.LedgerEntryVo) {
	cost := entry.Amount.Add(entry.Fee)
	position.shares += entry.Shares
	position.totalCost = position.totalCost.Add(cost)
	position.boughtShares += entry.Shares
	position.totalBuyCost = position.totalBuyCost.Add(cost)

	if entry.TradeDate >= position.latestBuyTradeDate {
		position.latestBuyTradeDate = entry.TradeDate
		position.name = entry.Name
		position.stopLossPrice = entry.StopLossPrice
		position.takeProfitPrice = entry.TakeProfitPrice
	}
}

// Sell releases cost at the average; the sale that empties the holding takes all remaining cost, so the
// holding's total realized profit is exact even though each partial sale's cost is rounded to the cent.
func (position *PositionDomain) Sell(entry vo.LedgerEntryVo, soldDate string) (decimal.Decimal, decimal.Decimal) {
	costOfSold := position.totalCost
	if entry.Shares < position.shares {
		costOfSold = position.totalCost.
			Mul(decimal.NewFromInt(int64(entry.Shares))).
			Div(decimal.NewFromInt(int64(position.shares))).
			Round(costOfSoldDecimalPlaces)
	}

	realized := entry.Amount.Sub(entry.Fee).Sub(costOfSold)
	position.shares = max(position.shares-entry.Shares, 0)
	position.totalCost = position.totalCost.Sub(costOfSold)
	position.realizedPnl = position.realizedPnl.Add(realized)
	if position.shares == 0 {
		position.closedDate = soldDate
	}

	return realized, costOfSold
}

func (position *PositionDomain) Shares() int {
	return position.shares
}

func (position *PositionDomain) IsClosed() bool {
	return position.closedDate != ""
}

func (position *PositionDomain) ClosedDate() string {
	return position.closedDate
}

func (position *PositionDomain) Result() vo.PositionResultVo {
	if position.realizedPnl.IsPositive() {
		return vo.PositionResultWin
	}

	return vo.PositionResultLoss
}

// ToDto shows an open holding as it stands today, and a closed one as everything that was bought and how it ended.
func (position *PositionDomain) ToDto(calendar BusinessCalendarDomain, today string) dto.PositionDto {
	if !position.IsClosed() {
		averageCost := position.totalCost.Div(decimal.NewFromInt(int64(position.shares))).Round(averageCostDecimalPlaces)
		positionDto := dto.PositionDto{
			Symbol:      position.symbol,
			Name:        position.name,
			Status:      string(vo.PositionStatusOpen),
			Shares:      position.shares,
			AverageCost: &averageCost,
			TotalCost:   position.totalCost,
			OpenedDate:  position.openedDate,
			// A fill dated after today is taken as given, but a holding is never held for negative days.
			DaysHeld:    max(calendar.DaysBetween(position.openedDate, today), 0),
			RealizedPnl: position.realizedPnl,
		}
		if position.stopLossPrice.Valid {
			positionDto.StopLossPrice = &position.stopLossPrice.Decimal
		}
		if position.takeProfitPrice.Valid {
			positionDto.TakeProfitPrice = &position.takeProfitPrice.Decimal
		}

		return positionDto
	}

	averageCost := position.totalBuyCost.Div(decimal.NewFromInt(int64(position.boughtShares))).Round(averageCostDecimalPlaces)
	closedDate := position.closedDate
	result := string(position.Result())

	return dto.PositionDto{
		Symbol:      position.symbol,
		Name:        position.name,
		Status:      string(vo.PositionStatusClosed),
		Shares:      position.boughtShares,
		AverageCost: &averageCost,
		TotalCost:   position.totalBuyCost,
		OpenedDate:  position.openedDate,
		ClosedDate:  &closedDate,
		DaysHeld:    calendar.DaysBetween(position.openedDate, position.closedDate),
		RealizedPnl: position.realizedPnl,
		Result:      &result,
	}
}
