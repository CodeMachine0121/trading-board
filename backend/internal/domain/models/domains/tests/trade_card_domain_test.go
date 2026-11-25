package domains_test

import (
	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func aBuyOrder() dto.BuyOrderWriteDto {
	return dto.BuyOrderWriteDto{
		Symbol: "2303", Name: "聯電", LimitPrice: money("52.5"), MaxChasePrice: money("54"), Shares: 90,
		EstimatedAmount: money("4725"), StopLossPrice: money("48"), TakeProfitPrice: money("60"), Reason: "突破",
	}
}

func aSellOrder() dto.SellOrderWriteDto {
	return dto.SellOrderWriteDto{Symbol: "3006", Name: "晶豪科", LimitPrice: money("68"), Shares: 50, ExitReason: "takeProfit"}
}

func aTradeCard(buyOrders []dto.BuyOrderWriteDto, sellOrders []dto.SellOrderWriteDto) dto.TradeCardPublishDto {
	return dto.TradeCardPublishDto{TradeDate: "2026-11-26", Status: "trade", BuyOrders: buyOrders, SellOrders: sellOrders}
}

func TestPublishedTradeCardDomainAcceptsAWellFormedCard(t *testing.T) {
	cases := []struct {
		name   string
		change func(*dto.BuyOrderWriteDto)
	}{
		{name: "stop 48 limit 52.5 chase 54 take 60", change: func(*dto.BuyOrderWriteDto) {}},
		{name: "limit equal to the highest chase price", change: func(order *dto.BuyOrderWriteDto) { order.LimitPrice = money("54") }},
		{name: "a single odd-lot share", change: func(order *dto.BuyOrderWriteDto) { order.Shares = 1 }},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			buyOrder := aBuyOrder()
			testCase.change(&buyOrder)

			cardDomain, publishError := domains.NewPublishedTradeCardDomain(
				aTradeCard([]dto.BuyOrderWriteDto{buyOrder}, nil), calendar, taipeiTime("2026-11-25 20:00"))

			require.NoError(t, publishError)
			cardDto := cardDomain.ToDto()
			require.Len(t, cardDto.BuyOrders, 1)
			assert.Equal(t, "pending", cardDto.BuyOrders[0].Status)
			assert.Equal(t, 0, cardDto.BuyOrders[0].FilledShares)
			assert.Empty(t, cardDto.BuyOrders[0].Fills)
			assert.Equal(t, buyOrder.Shares, cardDto.BuyOrders[0].Shares)
		})
	}
}

func TestPublishedTradeCardDomainRefusesABrokenBuyOrder(t *testing.T) {
	cases := []struct {
		name            string
		change          func(*dto.BuyOrderWriteDto)
		expectedMessage string
	}{
		{name: "stop-loss equal to the limit", change: func(order *dto.BuyOrderWriteDto) { order.StopLossPrice = money("52.5") }, expectedMessage: "停損價必須低於限價"},
		{name: "take-profit equal to the limit", change: func(order *dto.BuyOrderWriteDto) { order.TakeProfitPrice = money("52.5") }, expectedMessage: "停利價必須高於限價"},
		{name: "limit above the highest chase price", change: func(order *dto.BuyOrderWriteDto) { order.LimitPrice = money("55") }, expectedMessage: "限價不得高於追價上限"},
		{name: "zero shares", change: func(order *dto.BuyOrderWriteDto) { order.Shares = 0 }, expectedMessage: "股數必須是至少 1 的整數"},
		{name: "a zero price", change: func(order *dto.BuyOrderWriteDto) { order.LimitPrice = money("0") }, expectedMessage: "價格與預估金額都必須大於 0"},
		{name: "a zero estimated amount", change: func(order *dto.BuyOrderWriteDto) { order.EstimatedAmount = money("0") }, expectedMessage: "價格與預估金額都必須大於 0"},
		{name: "no stock code", change: func(order *dto.BuyOrderWriteDto) { order.Symbol = " " }, expectedMessage: "缺少股票代號"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			buyOrder := aBuyOrder()
			testCase.change(&buyOrder)

			_, publishError := domains.NewPublishedTradeCardDomain(
				aTradeCard([]dto.BuyOrderWriteDto{buyOrder}, nil), calendar, taipeiTime("2026-11-25 20:00"))

			require.ErrorIs(t, publishError, domains.ErrValidation)
			assert.Contains(t, publishError.Error(), testCase.expectedMessage)
		})
	}
}

func TestPublishedTradeCardDomainRefusesABrokenSellOrder(t *testing.T) {
	cases := []struct {
		name   string
		change func(*dto.SellOrderWriteDto)
	}{
		{name: "zero shares", change: func(order *dto.SellOrderWriteDto) { order.Shares = 0 }},
		{name: "a zero price", change: func(order *dto.SellOrderWriteDto) { order.LimitPrice = money("0") }},
		{name: "an unknown exit reason", change: func(order *dto.SellOrderWriteDto) { order.ExitReason = "panic" }},
		{name: "no stock code", change: func(order *dto.SellOrderWriteDto) { order.Symbol = "" }},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			sellOrder := aSellOrder()
			testCase.change(&sellOrder)

			_, publishError := domains.NewPublishedTradeCardDomain(
				aTradeCard(nil, []dto.SellOrderWriteDto{sellOrder}), calendar, taipeiTime("2026-11-25 20:00"))

			require.ErrorIs(t, publishError, domains.ErrValidation)
		})
	}
}

func TestPublishedTradeCardDomainKeepsTradeAndNoTradeApart(t *testing.T) {
	cases := []struct {
		name       string
		card       dto.TradeCardPublishDto
		isAccepted bool
	}{
		{name: "a no-trade card with a note and no orders", card: dto.TradeCardPublishDto{TradeDate: "2026-11-26", Status: "noTrade", Note: "大盤轉弱"}, isAccepted: true},
		{name: "a no-trade card carrying an order", card: dto.TradeCardPublishDto{TradeDate: "2026-11-26", Status: "noTrade", BuyOrders: []dto.BuyOrderWriteDto{aBuyOrder()}}, isAccepted: false},
		{name: "a trade card with no orders", card: dto.TradeCardPublishDto{TradeDate: "2026-11-26", Status: "trade"}, isAccepted: false},
		{name: "an unknown status", card: dto.TradeCardPublishDto{TradeDate: "2026-11-26", Status: "maybe"}, isAccepted: false},
		{name: "a session date that is not a date", card: dto.TradeCardPublishDto{TradeDate: "2026-11-31", Status: "noTrade"}, isAccepted: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			cardDomain, publishError := domains.NewPublishedTradeCardDomain(testCase.card, calendar, taipeiTime("2026-11-25 20:00"))

			if !testCase.isAccepted {
				require.ErrorIs(t, publishError, domains.ErrValidation)
				return
			}
			require.NoError(t, publishError)
			cardDto := cardDomain.ToDto()
			assert.Equal(t, "noTrade", cardDto.Status)
			assert.Equal(t, "大盤轉弱", cardDto.Note)
			assert.Empty(t, cardDto.BuyOrders)
			assert.Empty(t, cardDto.SellOrders)
		})
	}
}

func TestPublishedTradeCardDomainAddsUpSellsPerStock(t *testing.T) {
	cardDomain, publishError := domains.NewPublishedTradeCardDomain(
		aTradeCard([]dto.BuyOrderWriteDto{aBuyOrder()}, []dto.SellOrderWriteDto{aSellOrder(), aSellOrder()}),
		calendar, taipeiTime("2026-11-25 20:00"))

	require.NoError(t, publishError)
	assert.Equal(t, map[string]int{"3006": 100}, cardDomain.SoldSharesBySymbol())
	assert.Equal(t, "2026-11-26", cardDomain.TradeDate())
	assert.False(t, cardDomain.HasFills())
	assert.Len(t, cardDomain.ToEntity().Orders, 3)
}

func TestTradeCardDomainSummarisesItsOrders(t *testing.T) {
	cards := []struct {
		name     string
		card     entities.TradeCard
		expected dto.TradeCardSummaryDto
	}{
		{
			name: "two buys and a sell, none filled",
			card: entities.TradeCard{TradeDate: "2026-11-26", Status: "trade", Orders: []entities.TradeOrder{
				{Side: "buy", Shares: 90}, {Side: "buy", Shares: 10}, {Side: "sell", Shares: 50},
			}},
			expected: dto.TradeCardSummaryDto{TradeDate: "2026-11-26", Status: "trade", BuyCount: 2, SellCount: 1, FilledOrderCount: 0},
		},
		{
			name: "one buy fully filled",
			card: entities.TradeCard{TradeDate: "2026-11-25", Status: "trade", Orders: []entities.TradeOrder{
				{Side: "buy", Shares: 90, Fills: []entities.OrderFill{{Shares: 90}}},
			}},
			expected: dto.TradeCardSummaryDto{TradeDate: "2026-11-25", Status: "trade", BuyCount: 1, SellCount: 0, FilledOrderCount: 1},
		},
	}

	for _, testCase := range cards {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, domains.NewTradeCardDomain(testCase.card, calendar).ToSummaryDto())
		})
	}
}

func TestTradeCardDomainKnowsWhetherAnyOrderHasAFill(t *testing.T) {
	filled := entities.TradeCard{Orders: []entities.TradeOrder{{Shares: 90}, {Shares: 90, Fills: []entities.OrderFill{{Shares: 10}}}}}
	skippedOnly := entities.TradeCard{Orders: []entities.TradeOrder{{Shares: 90, IsSkipped: true}}}

	assert.True(t, domains.NewTradeCardDomain(filled, calendar).HasFills())
	assert.False(t, domains.NewTradeCardDomain(skippedOnly, calendar).HasFills())
}

func TestTradeCardDomainShowsSellOrdersWithTheirExitReason(t *testing.T) {
	card := entities.TradeCard{TradeDate: "2026-11-26", Status: "trade", PublishedAt: taipeiTime("2026-11-25 20:00").UTC(),
		Orders: []entities.TradeOrder{{Side: "sell", Symbol: "3006", Shares: 50, ExitReason: "stopLoss"}}}

	cardDto := domains.NewTradeCardDomain(card, calendar).ToDto()

	require.Len(t, cardDto.SellOrders, 1)
	assert.Equal(t, "stopLoss", *cardDto.SellOrders[0].ExitReason)
	assert.Nil(t, cardDto.SellOrders[0].StopLossPrice)
	assert.Nil(t, cardDto.SellOrders[0].IsMonthEndAllocation)
	assert.Equal(t, "2026-11-25T20:00:00+08:00", cardDto.PublishedAt.Format("2006-01-02T15:04:05Z07:00"))
}
