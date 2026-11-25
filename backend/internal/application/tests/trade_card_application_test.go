package application_test

import (
	"errors"
	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var errStorage = errors.New("storage unavailable")

func aPublishedBuy() dto.BuyOrderWriteDto {
	return dto.BuyOrderWriteDto{Symbol: "2303", Name: "聯電", LimitPrice: money("52.5"), MaxChasePrice: money("54"),
		Shares: 90, EstimatedAmount: money("4725"), StopLossPrice: money("48"), TakeProfitPrice: money("60")}
}

func aPublishedSell(shares int) dto.SellOrderWriteDto {
	return dto.SellOrderWriteDto{Symbol: "3006", Name: "晶豪科", LimitPrice: money("68"), Shares: shares, ExitReason: "takeProfit"}
}

func aPublishDto(sells ...dto.SellOrderWriteDto) dto.TradeCardPublishDto {
	return dto.TradeCardPublishDto{TradeDate: "2026-11-26", Status: "trade",
		BuyOrders: []dto.BuyOrderWriteDto{aPublishedBuy()}, SellOrders: sells}
}

// storedAs plays the database: it hands back what was created, with identities assigned.
func storedAs(card entities.TradeCard) entities.TradeCard {
	card.ID = "card-1"
	for index := range card.Orders {
		card.Orders[index].ID = "order-" + card.Orders[index].Symbol
		card.Orders[index].TradeCardID = card.ID
	}

	return card
}

func TestTradeCardApplicationPublishesANewCard(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00").withFillLog("")
	testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-26").Return(entities.TradeCard{}, false, nil)
	testHarness.cardRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, card entities.TradeCard) (entities.TradeCard, error) { return storedAs(card), nil })

	cardDto, publishError := testHarness.tradeCardApplication.PublishCard(t.Context(), aPublishDto())

	require.NoError(t, publishError)
	assert.Equal(t, "2026-11-26", cardDto.TradeDate)
	require.Len(t, cardDto.BuyOrders, 1)
	assert.Equal(t, "order-2303", cardDto.BuyOrders[0].ID)
	assert.Equal(t, "pending", cardDto.BuyOrders[0].Status)
	assert.Equal(t, 0, cardDto.BuyOrders[0].FilledShares)
	assert.Empty(t, cardDto.BuyOrders[0].Fills)
}

func TestTradeCardApplicationReplacesACardWithoutFills(t *testing.T) {
	cases := []struct {
		name     string
		existing entities.TradeCard
	}{
		{name: "earlier card has no fills", existing: entities.TradeCard{ID: "old", TradeDate: "2026-11-26",
			Orders: []entities.TradeOrder{{Symbol: "2330", Side: "buy", Shares: 10}}}},
		{name: "earlier card only has a skipped order", existing: entities.TradeCard{ID: "old", TradeDate: "2026-11-26",
			Orders: []entities.TradeOrder{{Symbol: "2330", Side: "buy", Shares: 10, IsSkipped: true}}}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00").withFillLog("2026-11-26")
			testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-26").Return(testCase.existing, true, nil)
			deleted := testHarness.cardRepository.EXPECT().Delete(gomock.Any(), "old").Return(nil)
			testHarness.cardRepository.EXPECT().Create(gomock.Any(), gomock.Any()).After(deleted).DoAndReturn(
				func(_ any, card entities.TradeCard) (entities.TradeCard, error) { return storedAs(card), nil })

			cardDto, publishError := testHarness.tradeCardApplication.PublishCard(t.Context(), aPublishDto())

			require.NoError(t, publishError)
			require.Len(t, cardDto.BuyOrders, 1)
			assert.Equal(t, "2303", cardDto.BuyOrders[0].Symbol)
		})
	}
}

func TestTradeCardApplicationRefusesToReplaceACardWithAFill(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")
	testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-26").Return(entities.TradeCard{
		ID: "old", Orders: []entities.TradeOrder{{Shares: 90, Fills: []entities.OrderFill{{Shares: 10}}}}}, true, nil)

	_, publishError := testHarness.tradeCardApplication.PublishCard(t.Context(), aPublishDto())

	require.ErrorIs(t, publishError, domains.ErrConflict)
	assert.Contains(t, publishError.Error(), "已有成交")
}

func TestTradeCardApplicationChecksSalesAgainstTheHolding(t *testing.T) {
	cases := []struct {
		name       string
		sells      []dto.SellOrderWriteDto
		isAccepted bool
	}{
		{name: "selling what is held", sells: []dto.SellOrderWriteDto{aPublishedSell(50)}, isAccepted: true},
		{name: "selling one share more than held", sells: []dto.SellOrderWriteDto{aPublishedSell(51)}, isAccepted: false},
		{name: "two sell orders together exceeding the holding", sells: []dto.SellOrderWriteDto{aPublishedSell(30), aPublishedSell(30)}, isAccepted: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00").
				withFillLog("2026-11-20", aHeldBuyOrder("3006", 50, "60", "2", "2026-11-20 09:05"))
			testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-26").Return(entities.TradeCard{}, false, nil)
			if testCase.isAccepted {
				testHarness.cardRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ any, card entities.TradeCard) (entities.TradeCard, error) { return storedAs(card), nil })
			}

			cardDto, publishError := testHarness.tradeCardApplication.PublishCard(t.Context(), aPublishDto(testCase.sells...))

			if !testCase.isAccepted {
				require.ErrorIs(t, publishError, domains.ErrValidation)
				assert.Contains(t, publishError.Error(), "超過持股")
				return
			}
			require.NoError(t, publishError)
			require.Len(t, cardDto.SellOrders, 1)
			assert.Equal(t, "pending", cardDto.SellOrders[0].Status)
		})
	}
}

func TestTradeCardApplicationRefusesAnInvalidCardBeforeTouchingStorage(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")
	invalid := aPublishDto()
	invalid.BuyOrders[0].StopLossPrice = money("52.5")

	_, publishError := testHarness.tradeCardApplication.PublishCard(t.Context(), invalid)

	require.ErrorIs(t, publishError, domains.ErrValidation)
}

func TestTradeCardApplicationPassesOnStorageFailuresWhilePublishing(t *testing.T) {
	cases := []struct {
		name    string
		arrange func(testHarness *harness)
	}{
		{name: "finding the earlier card", arrange: func(testHarness *harness) {
			testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), gomock.Any()).Return(entities.TradeCard{}, false, errStorage)
		}},
		{name: "reading the fill log", arrange: func(testHarness *harness) {
			testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), gomock.Any()).Return(entities.TradeCard{}, false, nil)
			testHarness.orderRepository.EXPECT().ListWithFills(gomock.Any()).Return(nil, errStorage)
		}},
		{name: "reading the earliest card", arrange: func(testHarness *harness) {
			testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), gomock.Any()).Return(entities.TradeCard{}, false, nil)
			testHarness.orderRepository.EXPECT().ListWithFills(gomock.Any()).Return(nil, nil)
			testHarness.cardRepository.EXPECT().FindEarliestTradeDate(gomock.Any()).Return("", false, errStorage)
		}},
		{name: "deleting the earlier card", arrange: func(testHarness *harness) {
			testHarness.withFillLog("2026-11-26")
			testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), gomock.Any()).Return(entities.TradeCard{ID: "old"}, true, nil)
			testHarness.cardRepository.EXPECT().Delete(gomock.Any(), "old").Return(errStorage)
		}},
		{name: "creating the card", arrange: func(testHarness *harness) {
			testHarness.withFillLog("")
			testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), gomock.Any()).Return(entities.TradeCard{}, false, nil)
			testHarness.cardRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(entities.TradeCard{}, errStorage)
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-25 20:00")
			testCase.arrange(testHarness)

			_, publishError := testHarness.tradeCardApplication.PublishCard(t.Context(), aPublishDto())

			assert.ErrorIs(t, publishError, errStorage)
		})
	}
}

func TestTradeCardApplicationReadsCards(t *testing.T) {
	stored := entities.TradeCard{ID: "c", TradeDate: "2026-11-26", Status: "trade", Orders: []entities.TradeOrder{
		{ID: "o", Side: "buy", Symbol: "2303", Shares: 90, Fills: []entities.OrderFill{{ID: "f", Shares: 40}}}}}

	t.Run("latest card is the one with the greatest session date", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")
		testHarness.cardRepository.EXPECT().FindLatest(gomock.Any()).Return(stored, true, nil)

		cardDto, findError := testHarness.tradeCardApplication.GetLatestCard(t.Context())

		require.NoError(t, findError)
		assert.Equal(t, "2026-11-26", cardDto.TradeDate)
		assert.Equal(t, "partiallyFilled", cardDto.BuyOrders[0].Status)
		assert.Len(t, cardDto.BuyOrders[0].Fills, 1)
	})

	t.Run("a card of a given date", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")
		testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-26").Return(stored, true, nil)

		cardDto, findError := testHarness.tradeCardApplication.GetCard(t.Context(), "2026-11-26")

		require.NoError(t, findError)
		assert.Equal(t, "2026-11-26", cardDto.TradeDate)
	})
}

func TestTradeCardApplicationAnswersNotFoundAndBadDates(t *testing.T) {
	t.Run("no card yet", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")
		testHarness.cardRepository.EXPECT().FindLatest(gomock.Any()).Return(entities.TradeCard{}, false, nil)

		_, findError := testHarness.tradeCardApplication.GetLatestCard(t.Context())

		assert.ErrorIs(t, findError, domains.ErrNotFound)
	})

	t.Run("a card of a date without one", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")
		testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-20").Return(entities.TradeCard{}, false, nil)

		_, findError := testHarness.tradeCardApplication.GetCard(t.Context(), "2026-11-20")

		assert.ErrorIs(t, findError, domains.ErrNotFound)
	})

	t.Run("a date that is not a date", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")

		_, findError := testHarness.tradeCardApplication.GetCard(t.Context(), "tomorrow")

		assert.ErrorIs(t, findError, domains.ErrValidation)
	})

	t.Run("storage failures pass on", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-25 20:00")
		testHarness.cardRepository.EXPECT().FindLatest(gomock.Any()).Return(entities.TradeCard{}, false, errStorage)
		testHarness.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), gomock.Any()).Return(entities.TradeCard{}, false, errStorage)
		testHarness.cardRepository.EXPECT().ListByMonth(gomock.Any(), gomock.Any()).Return(nil, errStorage)

		_, latestError := testHarness.tradeCardApplication.GetLatestCard(t.Context())
		_, cardError := testHarness.tradeCardApplication.GetCard(t.Context(), "2026-11-26")
		_, listError := testHarness.tradeCardApplication.ListCards(t.Context(), "2026-11")

		assert.ErrorIs(t, latestError, errStorage)
		assert.ErrorIs(t, cardError, errStorage)
		assert.ErrorIs(t, listError, errStorage)
	})
}

func TestTradeCardApplicationListsAMonthNewestFirst(t *testing.T) {
	cases := []struct {
		name          string
		month         string
		expectedMonth string
	}{
		{name: "the given month", month: "2026-11", expectedMonth: "2026-11"},
		{name: "the current month when none is given", month: "", expectedMonth: "2026-12"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-12-01 08:00")
			testHarness.cardRepository.EXPECT().ListByMonth(gomock.Any(), testCase.expectedMonth).Return([]entities.TradeCard{
				{TradeDate: "2026-11-26", Status: "trade", Orders: []entities.TradeOrder{
					{Side: "buy", Shares: 1}, {Side: "buy", Shares: 1}, {Side: "sell", Shares: 1}}},
				{TradeDate: "2026-11-25", Status: "trade", Orders: []entities.TradeOrder{
					{Side: "buy", Shares: 90, Fills: []entities.OrderFill{{Shares: 90}}}}},
			}, nil)

			summaries, listError := testHarness.tradeCardApplication.ListCards(t.Context(), testCase.month)

			require.NoError(t, listError)
			assert.Equal(t, []dto.TradeCardSummaryDto{
				{TradeDate: "2026-11-26", Status: "trade", BuyCount: 2, SellCount: 1, FilledOrderCount: 0},
				{TradeDate: "2026-11-25", Status: "trade", BuyCount: 1, SellCount: 0, FilledOrderCount: 1},
			}, summaries)
		})
	}
}

func TestTradeCardApplicationRefusesAnUnreadableMonth(t *testing.T) {
	testHarness := newHarness(t, "2026-12-01 08:00")

	_, listError := testHarness.tradeCardApplication.ListCards(t.Context(), "Nov")

	assert.ErrorIs(t, listError, domains.ErrValidation)
}
