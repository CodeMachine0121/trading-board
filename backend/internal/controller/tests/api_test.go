package controller_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const cardBody = `{"tradeDate":"2026-11-26","status":"trade","note":"note","buyOrders":[{"symbol":"2303","name":"聯電",
"limitPrice":"52.5","maxChasePrice":"54","shares":90,"estimatedAmount":"4725","stopLossPrice":"48",
"takeProfitPrice":"60","reason":"r","isMonthEndAllocation":false}],"sellOrders":[]}`

func TestHealthAnswersWithoutCredentials(t *testing.T) {
	testApi := newApi(t)

	for _, path := range []string{"/health", "/api/v1/health"} {
		result := testApi.call(t, http.MethodGet, path, "", nil)

		assert.Equal(t, http.StatusOK, result.status, path)
		assert.Equal(t, "ok", result.body["status"])
	}
}

func TestReadingRequiresAnAccessTokenOrAnActiveApiKey(t *testing.T) {
	cases := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
	}{
		{name: "no access token and no key", headers: nil, expectedStatus: http.StatusUnauthorized},
		{name: "an expired access token", headers: map[string]string{"Authorization": "Bearer expired"}, expectedStatus: http.StatusUnauthorized},
		{name: "an unknown or revoked key", headers: map[string]string{"X-Api-Key": "tbk_revoked"}, expectedStatus: http.StatusUnauthorized},
		{name: "a member's key", headers: asAgent(), expectedStatus: http.StatusOK},
		{name: "the owner's access token", headers: asOwner(), expectedStatus: http.StatusOK},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testApi := newApi(t)
			testApi.cardRepository.EXPECT().FindLatest(gomock.Any()).Return(storedCard(), true, nil).AnyTimes()

			result := testApi.call(t, http.MethodGet, "/api/v1/cards/latest", "", testCase.headers)

			assert.Equal(t, testCase.expectedStatus, result.status)
			if testCase.expectedStatus == http.StatusUnauthorized {
				assert.Equal(t, "unauthenticated", errorCodeOf(result))
				return
			}
			assert.Equal(t, "2026-11-26", result.body["tradeDate"])
			buyOrder := result.body["buyOrders"].([]any)[0].(map[string]any)
			assert.Equal(t, "order-1", buyOrder["id"])
		})
	}
}

func TestOnlyTheAgentKeyPublishes(t *testing.T) {
	testApi := newApi(t)

	asOwnerResult := testApi.call(t, http.MethodPut, "/api/v1/cards/2026-11-26", cardBody, asOwner())
	anonymous := testApi.call(t, http.MethodPut, "/api/v1/cards/2026-11-26", cardBody, nil)

	assert.Equal(t, http.StatusUnauthorized, asOwnerResult.status)
	assert.Equal(t, "unauthenticated", errorCodeOf(asOwnerResult))
	assert.Equal(t, http.StatusUnauthorized, anonymous.status)
}

func TestPublishingAnswersTheFullCardWithDecimalsAsStrings(t *testing.T) {
	testApi := newApi(t).withEmptyFillLog()
	testApi.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-26").Return(entities.TradeCard{}, false, nil)
	testApi.cardRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, card entities.TradeCard) (entities.TradeCard, error) {
			card.Orders[0].ID = "order-1"
			return card, nil
		})

	result := testApi.call(t, http.MethodPut, "/api/v1/cards/2026-11-26", cardBody, asAgent())

	require.Equal(t, http.StatusOK, result.status)
	assert.Equal(t, "2026-11-26", result.body["tradeDate"])
	assert.Equal(t, "trade", result.body["status"])
	assert.Equal(t, "2026-11-25T20:00:00+08:00", result.body["publishedAt"])
	assert.Equal(t, "desk-chief", result.body["publishedBy"])
	assert.Equal(t, []any{}, result.body["sellOrders"])
	buyOrder := result.body["buyOrders"].([]any)[0].(map[string]any)
	assert.Equal(t, "order-1", buyOrder["id"])
	assert.Equal(t, "52.5", buyOrder["limitPrice"])
	assert.Equal(t, "54", buyOrder["maxChasePrice"])
	assert.Equal(t, "4725", buyOrder["estimatedAmount"])
	assert.Equal(t, float64(90), buyOrder["shares"])
	assert.Equal(t, false, buyOrder["isMonthEndAllocation"])
	assert.Equal(t, "pending", buyOrder["status"])
	assert.Equal(t, float64(0), buyOrder["filledShares"])
	assert.Equal(t, []any{}, buyOrder["fills"])
}

func TestPublishingRefusesInvalidCards(t *testing.T) {
	cases := []struct {
		name         string
		path         string
		body         string
		expectedCode string
	}{
		{name: "stop-loss equal to the limit", path: "/api/v1/cards/2026-11-26",
			body: `{"status":"trade","buyOrders":[{"symbol":"2303","limitPrice":"52.5","maxChasePrice":"54","shares":90,
"estimatedAmount":"4725","stopLossPrice":"52.5","takeProfitPrice":"60"}]}`, expectedCode: "invalidCard"},
		{name: "a body date that differs from the path", path: "/api/v1/cards/2026-11-27", body: cardBody, expectedCode: "invalidRequest"},
		{name: "an unreadable body", path: "/api/v1/cards/2026-11-26", body: `{"buyOrders":"x"}`, expectedCode: "invalidRequest"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result := newApi(t).call(t, http.MethodPut, testCase.path, testCase.body, asAgent())

			assert.Equal(t, http.StatusBadRequest, result.status)
			assert.Equal(t, testCase.expectedCode, errorCodeOf(result))
		})
	}
}

func TestPublishingOverAFilledCardIsAConflict(t *testing.T) {
	testApi := newApi(t)
	filled := storedCard()
	filled.Orders[0].Fills = []entities.OrderFill{{ID: "fill-1", Shares: 10}}
	testApi.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-26").Return(filled, true, nil)

	result := testApi.call(t, http.MethodPut, "/api/v1/cards/2026-11-26", cardBody, asAgent())

	assert.Equal(t, http.StatusConflict, result.status)
	assert.Equal(t, "cardHasFills", errorCodeOf(result))
}

func TestReadingCards(t *testing.T) {
	testApi := newApi(t)
	testApi.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-26").Return(storedCard(), true, nil)
	testApi.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-20").Return(entities.TradeCard{}, false, nil)
	testApi.cardRepository.EXPECT().ListByMonth(gomock.Any(), "2026-11").Return([]entities.TradeCard{storedCard()}, nil)
	testApi.cardRepository.EXPECT().FindLatest(gomock.Any()).Return(entities.TradeCard{}, false, nil)

	found := testApi.call(t, http.MethodGet, "/api/v1/cards/2026-11-26", "", asOwner())
	missing := testApi.call(t, http.MethodGet, "/api/v1/cards/2026-11-20", "", asOwner())
	listed := testApi.call(t, http.MethodGet, "/api/v1/cards?month=2026-11", "", asOwner())
	noLatest := testApi.call(t, http.MethodGet, "/api/v1/cards/latest", "", asOwner())
	badMonth := testApi.call(t, http.MethodGet, "/api/v1/cards?month=11", "", asOwner())

	assert.Equal(t, http.StatusOK, found.status)
	assert.Equal(t, "2026-11-26", found.body["tradeDate"])
	assert.Equal(t, http.StatusNotFound, missing.status)
	assert.Equal(t, "cardNotFound", errorCodeOf(missing))
	require.Equal(t, http.StatusOK, listed.status)
	assert.Equal(t, []map[string]any{{"tradeDate": "2026-11-26", "status": "trade", "buyCount": float64(1),
		"sellCount": float64(0), "filledOrderCount": float64(0), "publishedBy": ""}}, listed.list)
	assert.Equal(t, http.StatusNotFound, noLatest.status)
	assert.Equal(t, http.StatusBadRequest, badMonth.status)
}

func TestReportingAFill(t *testing.T) {
	testApi := newApi(t).withEmptyFillLog()
	order := storedCard().Orders[0]
	first := testApi.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(order, true, nil)
	testApi.fillRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, fill entities.OrderFill) (entities.OrderFill, error) {
			fill.ID = "fill-1"
			order.Fills = append(order.Fills, fill)
			testApi.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").After(first).Return(order, true, nil)

			return fill, nil
		})

	result := testApi.call(t, http.MethodPost, "/api/v1/orders/order-1/fills",
		`{"price":"52.5","shares":90,"filledAt":"2026-11-26T09:05:00+08:00"}`, asOwner())

	require.Equal(t, http.StatusOK, result.status)
	assert.Equal(t, "filled", result.body["status"])
	assert.Equal(t, float64(90), result.body["filledShares"])
	fill := result.body["fills"].([]any)[0].(map[string]any)
	assert.Equal(t, map[string]any{"id": "fill-1", "price": "52.5", "shares": float64(90), "fee": "4", "amount": "4725",
		"filledAt": "2026-11-26T09:05:00+08:00", "reportedBy": "owner"}, fill)
}

func TestReportingAFillMapsRefusalsToStatuses(t *testing.T) {
	skipped := storedCard().Orders[0]
	skipped.IsSkipped = true

	cases := []struct {
		name           string
		arrange        func(testApi *api)
		body           string
		expectedStatus int
		expectedCode   string
	}{
		{name: "an unknown order", arrange: func(testApi *api) {
			testApi.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(entities.TradeOrder{}, false, nil)
		}, body: `{"price":"52.5","shares":1,"filledAt":"2026-11-26T09:05:00+08:00"}`, expectedStatus: http.StatusNotFound, expectedCode: "orderNotFound"},
		{name: "overfilling", arrange: func(testApi *api) {
			testApi.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(storedCard().Orders[0], true, nil)
		}, body: `{"price":"52.5","shares":91,"filledAt":"2026-11-26T09:05:00+08:00"}`, expectedStatus: http.StatusBadRequest, expectedCode: "fillExceedsOrder"},
		{name: "a skipped order", arrange: func(testApi *api) {
			testApi.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(skipped, true, nil)
		}, body: `{"price":"52.5","shares":1,"filledAt":"2026-11-26T09:05:00+08:00"}`, expectedStatus: http.StatusConflict, expectedCode: "orderSkipped"},
		{name: "an unreadable body", arrange: func(*api) {}, body: `{"filledAt":"yesterday"}`,
			expectedStatus: http.StatusBadRequest, expectedCode: "invalidRequest"},
		{name: "a storage outage", arrange: func(testApi *api) {
			testApi.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(entities.TradeOrder{}, false, errors.New("down"))
		}, body: `{"price":"52.5","shares":1,"filledAt":"2026-11-26T09:05:00+08:00"}`, expectedStatus: http.StatusInternalServerError, expectedCode: "internalError"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testApi := newApi(t)
			testCase.arrange(testApi)

			result := testApi.call(t, http.MethodPost, "/api/v1/orders/order-1/fills", testCase.body, asOwner())

			assert.Equal(t, testCase.expectedStatus, result.status)
			assert.Equal(t, testCase.expectedCode, errorCodeOf(result))
		})
	}
}

func TestSkippingAnOrder(t *testing.T) {
	testApi := newApi(t)
	filled := storedCard().Orders[0]
	filled.ID = "order-2"
	filled.Fills = []entities.OrderFill{{ID: "f", Shares: 90}}
	testApi.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(storedCard().Orders[0], true, nil)
	testApi.orderRepository.EXPECT().FindByID(gomock.Any(), "order-2").Return(filled, true, nil)
	testApi.orderRepository.EXPECT().MarkSkipped(gomock.Any(), "order-1", "desk-chief").Return(nil)

	skipped := testApi.call(t, http.MethodPost, "/api/v1/orders/order-1/skip", "", asAgent())
	refused := testApi.call(t, http.MethodPost, "/api/v1/orders/order-2/skip", "", asOwner())

	assert.Equal(t, http.StatusOK, skipped.status)
	assert.Equal(t, "skipped", skipped.body["status"])
	assert.Equal(t, "desk-chief", skipped.body["skippedBy"])
	assert.Equal(t, http.StatusConflict, refused.status)
	assert.Equal(t, "orderAlreadyFilled", errorCodeOf(refused))
}

func TestReadingPositionsAndTheMonthlySummary(t *testing.T) {
	testApi := newApi(t).withEmptyFillLog()

	positions := testApi.call(t, http.MethodGet, "/api/v1/positions?status=closed&month=2026-11", "", asOwner())
	badStatus := testApi.call(t, http.MethodGet, "/api/v1/positions?status=sold", "", asOwner())
	summary := testApi.call(t, http.MethodGet, "/api/v1/months/2026-11/summary", "", asOwner())
	beforeFirst := testApi.call(t, http.MethodGet, "/api/v1/months/2026-10/summary", "", asOwner())

	assert.Equal(t, http.StatusOK, positions.status)
	assert.Empty(t, positions.list)
	assert.Equal(t, http.StatusBadRequest, badStatus.status)
	require.Equal(t, http.StatusOK, summary.status)
	assert.Equal(t, map[string]any{"month": "2026-11", "mode": "monthEnd", "allocation": "10000", "carriedOver": "0",
		"recycledProceeds": "0", "spent": "0", "remaining": "10000", "realizedPnl": "0", "closedPositionCount": float64(0),
		"winCount": float64(0), "lossCount": float64(0), "winRate": "0", "realizedReturnPct": "0"}, summary.body)
	assert.Equal(t, http.StatusNotFound, beforeFirst.status)
	assert.Equal(t, "monthNotFound", errorCodeOf(beforeFirst))
}

func TestOpenPositionsAreAJsonArray(t *testing.T) {
	result := newApi(t).withEmptyFillLog().call(t, http.MethodGet, "/api/v1/positions", "", asAgent())

	require.Equal(t, http.StatusOK, result.status)
	assert.NotNil(t, result.list)
}

func TestCorsLetsTheConfiguredFrontendRead(t *testing.T) {
	testApi := newApi(t)

	preflight := testApi.call(t, http.MethodOptions, "/api/v1/cards/latest", "",
		map[string]string{"Origin": "https://trading-board.coding-afternoon.com"})
	stranger := testApi.call(t, http.MethodGet, "/health", "", map[string]string{"Origin": "https://evil.example"})

	assert.Equal(t, http.StatusNoContent, preflight.status)
	assert.Equal(t, "https://trading-board.coding-afternoon.com", preflight.header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, preflight.header.Get("Access-Control-Allow-Headers"), "X-Api-Key")
	assert.Equal(t, "true", preflight.header.Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, stranger.header.Get("Access-Control-Allow-Origin"))
}

func TestPublishingASellOrderOfAHolding(t *testing.T) {
	testApi := newApi(t)
	held := storedCard().Orders[0]
	held.Symbol = "3006"
	held.Fills = []entities.OrderFill{{ID: "f", Shares: 50, Price: decimal.RequireFromString("60"),
		Amount: decimal.RequireFromString("3000"), Fee: decimal.RequireFromString("2"),
		FilledAt: time.Date(2026, 11, 20, 9, 5, 0, 0, taipei), RecordedAt: time.Date(2026, 11, 20, 9, 5, 0, 0, taipei)}}
	testApi.orderRepository.EXPECT().ListWithFills(gomock.Any()).Return([]entities.TradeOrder{held}, nil)
	testApi.cardRepository.EXPECT().FindEarliestTradeDate(gomock.Any()).Return("2026-11-20", true, nil)
	testApi.cardRepository.EXPECT().FindByTradeDate(gomock.Any(), "2026-11-26").Return(entities.TradeCard{}, false, nil)
	testApi.cardRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, card entities.TradeCard) (entities.TradeCard, error) { return card, nil })

	result := testApi.call(t, http.MethodPut, "/api/v1/cards/2026-11-26", `{"status":"trade","sellOrders":[{"symbol":"3006",
"name":"晶豪科","limitPrice":"68","shares":50,"exitReason":"takeProfit","reason":"到價"}]}`, asAgent())

	require.Equal(t, http.StatusOK, result.status)
	sellOrder := result.body["sellOrders"].([]any)[0].(map[string]any)
	assert.Equal(t, "takeProfit", sellOrder["exitReason"])
	assert.Equal(t, "68", sellOrder["limitPrice"])
	assert.NotContains(t, sellOrder, "stopLossPrice")
}
