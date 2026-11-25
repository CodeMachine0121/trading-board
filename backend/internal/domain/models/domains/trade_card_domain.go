package domains

import (
	"fmt"
	"strings"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/shopspring/decimal"
)

const invalidCardCode = "invalidCard"

var knownExitReasons = map[vo.ExitReasonVo]bool{
	vo.ExitReasonTakeProfit:   true,
	vo.ExitReasonStopLoss:     true,
	vo.ExitReasonTrailingStop: true,
	vo.ExitReasonTimeStop:     true,
}

// TradeCardDomain is a Daily Trade Card: the publish rules on the way in and its views on the way out.
type TradeCardDomain struct {
	card     entities.TradeCard
	calendar BusinessCalendarDomain
}

func NewTradeCardDomain(card entities.TradeCard, calendar BusinessCalendarDomain) TradeCardDomain {
	return TradeCardDomain{card: card, calendar: calendar}
}

// NewPublishedTradeCardDomain checks every rule a published card must meet; the first broken rule is named.
func NewPublishedTradeCardDomain(
	publishDto dto.TradeCardPublishDto, calendar BusinessCalendarDomain, publishedAt time.Time,
) (TradeCardDomain, error) {
	if !calendar.IsDate(publishDto.TradeDate) {
		return TradeCardDomain{}, ValidationFailed(invalidCardCode, "交易日須為 YYYY-MM-DD")
	}

	orderCount := len(publishDto.BuyOrders) + len(publishDto.SellOrders)
	switch vo.CardStatusVo(publishDto.Status) {
	case vo.CardStatusTrade:
		if orderCount == 0 {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, "有交易的交易卡至少要有一筆委託")
		}
	case vo.CardStatusNoTrade:
		if orderCount > 0 {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, "今日不買的交易卡不能帶任何委託")
		}
	default:
		return TradeCardDomain{}, ValidationFailed(invalidCardCode, "交易卡狀態只能是 trade 或 noTrade")
	}

	card := entities.TradeCard{
		TradeDate:   publishDto.TradeDate,
		Status:      publishDto.Status,
		Note:        publishDto.Note,
		PublishedAt: publishedAt,
		PublishedBy: publishDto.PublishedBy,
		Orders:      make([]entities.TradeOrder, 0, orderCount),
	}

	for index, buyOrder := range publishDto.BuyOrders {
		position := fmt.Sprintf("第 %d 筆買進", index+1)
		if strings.TrimSpace(buyOrder.Symbol) == "" {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, position+"缺少股票代號")
		}
		if !buyOrder.LimitPrice.IsPositive() || !buyOrder.MaxChasePrice.IsPositive() ||
			!buyOrder.StopLossPrice.IsPositive() || !buyOrder.TakeProfitPrice.IsPositive() ||
			!buyOrder.EstimatedAmount.IsPositive() {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, position+"的價格與預估金額都必須大於 0")
		}
		if buyOrder.Shares < 1 {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, position+"的股數必須是至少 1 的整數")
		}
		if !buyOrder.StopLossPrice.LessThan(buyOrder.LimitPrice) {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, position+"的停損價必須低於限價")
		}
		if !buyOrder.TakeProfitPrice.GreaterThan(buyOrder.LimitPrice) {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, position+"的停利價必須高於限價")
		}
		if buyOrder.LimitPrice.GreaterThan(buyOrder.MaxChasePrice) {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, position+"的限價不得高於追價上限")
		}

		card.Orders = append(card.Orders, entities.TradeOrder{
			TradeDate:            publishDto.TradeDate,
			Side:                 string(vo.OrderSideBuy),
			Sequence:             len(card.Orders),
			Symbol:               strings.TrimSpace(buyOrder.Symbol),
			Name:                 buyOrder.Name,
			LimitPrice:           buyOrder.LimitPrice,
			MaxChasePrice:        decimal.NewNullDecimal(buyOrder.MaxChasePrice),
			Shares:               buyOrder.Shares,
			EstimatedAmount:      decimal.NewNullDecimal(buyOrder.EstimatedAmount),
			StopLossPrice:        decimal.NewNullDecimal(buyOrder.StopLossPrice),
			TakeProfitPrice:      decimal.NewNullDecimal(buyOrder.TakeProfitPrice),
			Reason:               buyOrder.Reason,
			IsMonthEndAllocation: buyOrder.IsMonthEndAllocation,
		})
	}

	for index, sellOrder := range publishDto.SellOrders {
		position := fmt.Sprintf("第 %d 筆賣出", index+1)
		if strings.TrimSpace(sellOrder.Symbol) == "" {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, position+"缺少股票代號")
		}
		if !sellOrder.LimitPrice.IsPositive() {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, position+"的限價必須大於 0")
		}
		if sellOrder.Shares < 1 {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode, position+"的股數必須是至少 1 的整數")
		}
		if !knownExitReasons[vo.ExitReasonVo(sellOrder.ExitReason)] {
			return TradeCardDomain{}, ValidationFailed(invalidCardCode,
				position+"的出場原因只能是 takeProfit、stopLoss、trailingStop 或 timeStop")
		}

		card.Orders = append(card.Orders, entities.TradeOrder{
			TradeDate:  publishDto.TradeDate,
			Side:       string(vo.OrderSideSell),
			Sequence:   len(card.Orders),
			Symbol:     strings.TrimSpace(sellOrder.Symbol),
			Name:       sellOrder.Name,
			LimitPrice: sellOrder.LimitPrice,
			Shares:     sellOrder.Shares,
			ExitReason: sellOrder.ExitReason,
			Reason:     sellOrder.Reason,
		})
	}

	return TradeCardDomain{card: card, calendar: calendar}, nil
}

func (cardDomain TradeCardDomain) TradeDate() string {
	return cardDomain.card.TradeDate
}

func (cardDomain TradeCardDomain) HasFills() bool {
	for _, order := range cardDomain.card.Orders {
		if len(order.Fills) > 0 {
			return true
		}
	}

	return false
}

// SoldSharesBySymbol adds up every sell order of the card per stock, since together they may not exceed the holding.
func (cardDomain TradeCardDomain) SoldSharesBySymbol() map[string]int {
	soldShares := map[string]int{}
	for _, order := range cardDomain.card.Orders {
		if vo.OrderSideVo(order.Side) == vo.OrderSideSell {
			soldShares[order.Symbol] += order.Shares
		}
	}

	return soldShares
}

func (cardDomain TradeCardDomain) ToEntity() entities.TradeCard {
	return cardDomain.card
}

func (cardDomain TradeCardDomain) ToDto() dto.TradeCardDto {
	cardDto := dto.TradeCardDto{
		TradeDate:   cardDomain.card.TradeDate,
		Status:      cardDomain.card.Status,
		Note:        cardDomain.card.Note,
		BuyOrders:   []dto.OrderDto{},
		SellOrders:  []dto.OrderDto{},
		PublishedAt: cardDomain.card.PublishedAt.In(cardDomain.calendar.Location()),
		PublishedBy: cardDomain.card.PublishedBy,
	}

	for _, order := range cardDomain.card.Orders {
		orderDto := NewTradeOrderDomain(order, cardDomain.calendar).ToDto()
		if vo.OrderSideVo(order.Side) == vo.OrderSideSell {
			cardDto.SellOrders = append(cardDto.SellOrders, orderDto)
			continue
		}
		cardDto.BuyOrders = append(cardDto.BuyOrders, orderDto)
	}

	return cardDto
}

func (cardDomain TradeCardDomain) ToSummaryDto() dto.TradeCardSummaryDto {
	summaryDto := dto.TradeCardSummaryDto{
		TradeDate: cardDomain.card.TradeDate, Status: cardDomain.card.Status, PublishedBy: cardDomain.card.PublishedBy,
	}
	for _, order := range cardDomain.card.Orders {
		if vo.OrderSideVo(order.Side) == vo.OrderSideSell {
			summaryDto.SellCount++
		} else {
			summaryDto.BuyCount++
		}
		if NewTradeOrderDomain(order, cardDomain.calendar).Status() == vo.OrderStatusFilled {
			summaryDto.FilledOrderCount++
		}
	}

	return summaryDto
}
