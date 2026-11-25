package dto

import "github.com/shopspring/decimal"

type BuyOrderWriteDto struct {
	Symbol               string
	Name                 string
	LimitPrice           decimal.Decimal
	MaxChasePrice        decimal.Decimal
	Shares               int
	EstimatedAmount      decimal.Decimal
	StopLossPrice        decimal.Decimal
	TakeProfitPrice      decimal.Decimal
	Reason               string
	IsMonthEndAllocation bool
}
