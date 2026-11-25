package domains

import (
	"fmt"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/shopspring/decimal"
)

// TradeOrderDomain owns an order's progress: what it may still accept and what its status is.
type TradeOrderDomain struct {
	order    entities.TradeOrder
	calendar BusinessCalendarDomain
}

func NewTradeOrderDomain(order entities.TradeOrder, calendar BusinessCalendarDomain) TradeOrderDomain {
	return TradeOrderDomain{order: order, calendar: calendar}
}

func (orderDomain TradeOrderDomain) FilledShares() int {
	filledShares := 0
	for _, fill := range orderDomain.order.Fills {
		filledShares += fill.Shares
	}

	return filledShares
}

// Status is derived from filled shares, except skipped, which the owner set and which wins.
func (orderDomain TradeOrderDomain) Status() vo.OrderStatusVo {
	filledShares := orderDomain.FilledShares()
	switch {
	case orderDomain.order.IsSkipped:
		return vo.OrderStatusSkipped
	case filledShares >= orderDomain.order.Shares:
		return vo.OrderStatusFilled
	case filledShares > 0:
		return vo.OrderStatusPartiallyFilled
	default:
		return vo.OrderStatusPending
	}
}

// ReportFill returns the new fill to store; the fee is estimated when the owner left it out.
func (orderDomain *TradeOrderDomain) ReportFill(
	report dto.OrderFillReportDto, feeSchedule FeeScheduleDomain, recordedAt time.Time,
) (entities.OrderFill, error) {
	if orderDomain.order.IsSkipped {
		return entities.OrderFill{}, Conflict("orderSkipped", "這筆委託已略過，不能再回報成交")
	}
	if !report.Price.IsPositive() || report.Shares < 1 || report.FilledAt.IsZero() ||
		(report.Fee.Valid && report.Fee.Decimal.IsNegative()) {
		return entities.OrderFill{}, ValidationFailed("invalidFill",
			"成交價須大於 0、股數須為至少 1 的整數、手續費不得為負，並須填寫成交時間")
	}

	remainingShares := orderDomain.order.Shares - orderDomain.FilledShares()
	if report.Shares > remainingShares {
		return entities.OrderFill{}, ValidationFailed("fillExceedsOrder",
			fmt.Sprintf("成交股數 %d 超過這筆委託尚未成交的 %d 股", report.Shares, remainingShares))
	}

	amount := report.Price.Mul(decimal.NewFromInt(int64(report.Shares)))
	fee := report.Fee.Decimal
	if !report.Fee.Valid {
		fee = feeSchedule.EstimateFee(vo.OrderSideVo(orderDomain.order.Side), amount)
	}

	fill := entities.OrderFill{
		TradeOrderID: orderDomain.order.ID,
		Price:        report.Price,
		Shares:       report.Shares,
		Fee:          fee,
		Amount:       amount,
		FilledAt:     report.FilledAt,
		RecordedAt:   recordedAt,
	}
	orderDomain.order.Fills = append(orderDomain.order.Fills, fill)

	return fill, nil
}

// Skip refuses only a fully filled order; skipping one already skipped changes nothing.
func (orderDomain *TradeOrderDomain) Skip() error {
	if !orderDomain.order.IsSkipped && orderDomain.Status() == vo.OrderStatusFilled {
		return Conflict("orderAlreadyFilled", "這筆委託已全部成交，不能略過")
	}
	orderDomain.order.IsSkipped = true

	return nil
}

// LedgerEntryOfLatestFill is the entry the ledger must admit for the fill just reported.
func (orderDomain TradeOrderDomain) LedgerEntryOfLatestFill() vo.LedgerEntryVo {
	return orderDomain.order.ToLedgerEntryVo(orderDomain.order.Fills[len(orderDomain.order.Fills)-1])
}

func (orderDomain TradeOrderDomain) ToDto() dto.OrderDto {
	fillDtos := make([]dto.OrderFillDto, 0, len(orderDomain.order.Fills))
	for _, fill := range orderDomain.order.Fills {
		fillDtos = append(fillDtos, fill.ToDto(orderDomain.calendar.Location()))
	}

	orderDto := dto.OrderDto{
		ID:           orderDomain.order.ID,
		Side:         orderDomain.order.Side,
		Symbol:       orderDomain.order.Symbol,
		Name:         orderDomain.order.Name,
		LimitPrice:   orderDomain.order.LimitPrice,
		Shares:       orderDomain.order.Shares,
		Reason:       orderDomain.order.Reason,
		Status:       string(orderDomain.Status()),
		FilledShares: orderDomain.FilledShares(),
		Fills:        fillDtos,
	}

	if vo.OrderSideVo(orderDomain.order.Side) == vo.OrderSideSell {
		exitReason := orderDomain.order.ExitReason
		orderDto.ExitReason = &exitReason

		return orderDto
	}

	maxChasePrice := orderDomain.order.MaxChasePrice.Decimal
	estimatedAmount := orderDomain.order.EstimatedAmount.Decimal
	stopLossPrice := orderDomain.order.StopLossPrice.Decimal
	takeProfitPrice := orderDomain.order.TakeProfitPrice.Decimal
	isMonthEndAllocation := orderDomain.order.IsMonthEndAllocation
	orderDto.MaxChasePrice = &maxChasePrice
	orderDto.EstimatedAmount = &estimatedAmount
	orderDto.StopLossPrice = &stopLossPrice
	orderDto.TakeProfitPrice = &takeProfitPrice
	orderDto.IsMonthEndAllocation = &isMonthEndAllocation

	return orderDto
}
