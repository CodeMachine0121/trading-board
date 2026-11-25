package dto

import "github.com/shopspring/decimal"

// PositionDto is one holding; fields that only make sense for open or for closed holdings are null otherwise.
type PositionDto struct {
	Symbol          string           `json:"symbol"`
	Name            string           `json:"name"`
	Status          string           `json:"status"`
	Shares          int              `json:"shares"`
	AverageCost     *decimal.Decimal `json:"averageCost"`
	TotalCost       decimal.Decimal  `json:"totalCost"`
	StopLossPrice   *decimal.Decimal `json:"stopLossPrice"`
	TakeProfitPrice *decimal.Decimal `json:"takeProfitPrice"`
	OpenedDate      string           `json:"openedDate"`
	ClosedDate      *string          `json:"closedDate"`
	DaysHeld        int              `json:"daysHeld"`
	RealizedPnl     decimal.Decimal  `json:"realizedPnl"`
	Result          *string          `json:"result"`
}
