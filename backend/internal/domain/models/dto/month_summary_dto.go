package dto

import "github.com/shopspring/decimal"

// MonthSummaryDto is the month's budget and performance, flat.
type MonthSummaryDto struct {
	Month               string          `json:"month"`
	Mode                string          `json:"mode"`
	Allocation          decimal.Decimal `json:"allocation"`
	CarriedOver         decimal.Decimal `json:"carriedOver"`
	RecycledProceeds    decimal.Decimal `json:"recycledProceeds"`
	Spent               decimal.Decimal `json:"spent"`
	Remaining           decimal.Decimal `json:"remaining"`
	RealizedPnl         decimal.Decimal `json:"realizedPnl"`
	ClosedPositionCount int             `json:"closedPositionCount"`
	WinCount            int             `json:"winCount"`
	LossCount           int             `json:"lossCount"`
	WinRate             decimal.Decimal `json:"winRate"`
	RealizedReturnPct   decimal.Decimal `json:"realizedReturnPct"`
}
