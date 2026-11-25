package domains

import (
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/shopspring/decimal"
)

// FeeScheduleDomain estimates the fee of a fill the owner reported without one.
type FeeScheduleDomain struct {
	schedule vo.FeeScheduleVo
}

// NewFeeScheduleDomain treats a negative setting as zero, since a negative fee would pay the owner to trade.
func NewFeeScheduleDomain(schedule vo.FeeScheduleVo) FeeScheduleDomain {
	schedule.BrokerFeeRate = decimal.Max(schedule.BrokerFeeRate, decimal.Zero)
	schedule.Discount = decimal.Max(schedule.Discount, decimal.Zero)
	schedule.MinimumFee = decimal.Max(schedule.MinimumFee, decimal.Zero)
	schedule.TransactionTaxRate = decimal.Max(schedule.TransactionTaxRate, decimal.Zero)

	return FeeScheduleDomain{schedule: schedule}
}

// EstimateFee: broker fee rounded to whole dollars and never below the minimum; a sell adds the transaction tax.
func (feeSchedule FeeScheduleDomain) EstimateFee(side vo.OrderSideVo, amount decimal.Decimal) decimal.Decimal {
	brokerFee := decimal.Max(
		amount.Mul(feeSchedule.schedule.BrokerFeeRate).Mul(feeSchedule.schedule.Discount).Round(0),
		feeSchedule.schedule.MinimumFee,
	)
	if side != vo.OrderSideSell {
		return brokerFee
	}

	return brokerFee.Add(amount.Mul(feeSchedule.schedule.TransactionTaxRate).Round(0))
}
