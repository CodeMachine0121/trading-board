package dto

import "github.com/shopspring/decimal"

type SellOrderWriteDto struct {
	Symbol     string
	Name       string
	LimitPrice decimal.Decimal
	Shares     int
	ExitReason string
	Reason     string
}
