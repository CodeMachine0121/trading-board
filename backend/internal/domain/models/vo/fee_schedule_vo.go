package vo

import "github.com/shopspring/decimal"

// FeeScheduleVo holds the operator's fee settings used when the owner leaves a fill's fee out.
type FeeScheduleVo struct {
	BrokerFeeRate      decimal.Decimal
	Discount           decimal.Decimal
	MinimumFee         decimal.Decimal
	TransactionTaxRate decimal.Decimal
}
