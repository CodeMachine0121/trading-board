package application_test

import (
	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func aWaitingBuyOrder(shares int, filled ...int) entities.TradeOrder {
	order := entities.TradeOrder{ID: "order-1", Side: "buy", Symbol: "2303", Name: "聯電", TradeDate: "2026-11-26",
		Shares: shares, StopLossPrice: decimal.NewNullDecimal(money("48")), TakeProfitPrice: decimal.NewNullDecimal(money("60"))}
	for index, shares := range filled {
		order.Fills = append(order.Fills, entities.OrderFill{ID: "earlier-" + string(rune('a'+index)), Shares: shares,
			Price: money("52.5"), Amount: money("52.5").Mul(decimal.NewFromInt(int64(shares))), Fee: money("1"),
			FilledAt: taipeiTime("2026-11-26 09:01"), RecordedAt: taipeiTime("2026-11-26 09:01")})
	}

	return order
}

func aFillReport(shares int, price string) dto.OrderFillReportDto {
	return dto.OrderFillReportDto{Price: money(price), Shares: shares, FilledAt: taipeiTime("2026-11-26 09:05")}
}

// storesFill makes the fill repository accept the fill and the order read back with it appended.
func (testHarness *harness) storesFill(order entities.TradeOrder) *entities.OrderFill {
	stored := &entities.OrderFill{}
	testHarness.fillRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, fill entities.OrderFill) (entities.OrderFill, error) {
			fill.ID = "new-fill"
			*stored = fill
			order.Fills = append(order.Fills, fill)
			testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), order.ID).Return(order, true, nil)

			return fill, nil
		})

	return stored
}

func TestTradeOrderApplicationRecordsFillProgress(t *testing.T) {
	cases := []struct {
		name           string
		alreadyFilled  []int
		reported       int
		expectedStatus string
		expectedFilled int
		expectedFills  int
	}{
		{name: "a full fill completes the order", reported: 90, expectedStatus: "filled", expectedFilled: 90, expectedFills: 1},
		{name: "a partial fill", reported: 40, expectedStatus: "partiallyFilled", expectedFilled: 40, expectedFills: 1},
		{name: "the last shares complete the order", alreadyFilled: []int{40}, reported: 50, expectedStatus: "filled", expectedFilled: 90, expectedFills: 2},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			order := aWaitingBuyOrder(90, testCase.alreadyFilled...)
			testHarness := newHarness(t, "2026-11-26 13:00").withFillLog("2026-11-26")
			testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(order, true, nil)
			testHarness.storesFill(order)

			orderDto, reportError := testHarness.tradeOrderApplication.ReportFill(t.Context(), "order-1",
				aFillReport(testCase.reported, "52.5"))

			require.NoError(t, reportError)
			assert.Equal(t, testCase.expectedStatus, orderDto.Status)
			assert.Equal(t, testCase.expectedFilled, orderDto.FilledShares)
			assert.Len(t, orderDto.Fills, testCase.expectedFills)
		})
	}
}

func TestTradeOrderApplicationEstimatesTheFeeItStores(t *testing.T) {
	testHarness := newHarness(t, "2026-11-26 13:00").withFillLog("2026-11-26")
	testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(aWaitingBuyOrder(90), true, nil)
	stored := testHarness.storesFill(aWaitingBuyOrder(90))

	orderDto, reportError := testHarness.tradeOrderApplication.ReportFill(t.Context(), "order-1", aFillReport(90, "52.5"))

	require.NoError(t, reportError)
	assert.Equal(t, "4725", stored.Amount.String())
	assert.Equal(t, "4", stored.Fee.String())
	assert.Equal(t, "new-fill", orderDto.Fills[0].ID)
	assert.Equal(t, "4", orderDto.Fills[0].Fee.String())
}

func TestTradeOrderApplicationRefusesFillsWithoutStoringThem(t *testing.T) {
	skipped := aWaitingBuyOrder(90)
	skipped.IsSkipped = true
	sellOrder := entities.TradeOrder{ID: "order-1", Side: "sell", Symbol: "3006", TradeDate: "2026-11-26", Shares: 60}

	cases := []struct {
		name         string
		order        entities.TradeOrder
		fillLog      []entities.TradeOrder
		report       dto.OrderFillReportDto
		expectedKind error
		expectedText string
	}{
		{name: "overfilling", order: aWaitingBuyOrder(90, 40), report: aFillReport(51, "52.5"), expectedKind: domains.ErrValidation, expectedText: "超過"},
		{name: "a skipped order takes no fills", order: skipped, report: aFillReport(10, "52.5"), expectedKind: domains.ErrConflict, expectedText: "已略過"},
		{name: "selling more than held", order: sellOrder,
			fillLog: []entities.TradeOrder{aHeldBuyOrder("3006", 50, "60", "2", "2026-11-20 09:05")},
			report:  aFillReport(60, "68"), expectedKind: domains.ErrConflict, expectedText: "只持有 50 股"},
		{name: "a buy that would overspend", order: aWaitingBuyOrder(90),
			fillLog: []entities.TradeOrder{aHeldBuyOrder("2330", 1, "5271", "1", "2026-11-20 09:05")},
			report:  aFillReport(90, "52.5"), expectedKind: domains.ErrConflict, expectedText: "預算不足"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-26 13:00").withFillLog("2026-11-20", testCase.fillLog...)
			testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(testCase.order, true, nil)

			_, reportError := testHarness.tradeOrderApplication.ReportFill(t.Context(), "order-1", testCase.report)

			require.ErrorIs(t, reportError, testCase.expectedKind)
			assert.Contains(t, reportError.Error(), testCase.expectedText)
		})
	}
}

func TestTradeOrderApplicationAnswersNotFoundForAnUnknownOrder(t *testing.T) {
	testHarness := newHarness(t, "2026-11-26 13:00")
	testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), "missing").Return(entities.TradeOrder{}, false, nil).Times(2)

	_, reportError := testHarness.tradeOrderApplication.ReportFill(t.Context(), "missing", aFillReport(10, "52.5"))
	_, skipError := testHarness.tradeOrderApplication.SkipOrder(t.Context(), "missing")

	assert.ErrorIs(t, reportError, domains.ErrNotFound)
	assert.ErrorIs(t, skipError, domains.ErrNotFound)
}

func TestTradeOrderApplicationPassesOnStorageFailuresWhileReporting(t *testing.T) {
	cases := []struct {
		name    string
		arrange func(testHarness *harness)
	}{
		{name: "finding the order", arrange: func(testHarness *harness) {
			testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(entities.TradeOrder{}, false, errStorage)
		}},
		{name: "reading the fill log", arrange: func(testHarness *harness) {
			testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(aWaitingBuyOrder(90), true, nil)
			testHarness.orderRepository.EXPECT().ListWithFills(gomock.Any()).Return(nil, errStorage)
		}},
		{name: "storing the fill", arrange: func(testHarness *harness) {
			testHarness.withFillLog("2026-11-26")
			testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(aWaitingBuyOrder(90), true, nil)
			testHarness.fillRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(entities.OrderFill{}, errStorage)
		}},
		{name: "reading the order back", arrange: func(testHarness *harness) {
			testHarness.withFillLog("2026-11-26")
			first := testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(aWaitingBuyOrder(90), true, nil)
			testHarness.fillRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(entities.OrderFill{}, nil)
			testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).After(first).Return(entities.TradeOrder{}, false, errStorage)
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-26 13:00")
			testCase.arrange(testHarness)

			_, reportError := testHarness.tradeOrderApplication.ReportFill(t.Context(), "order-1", aFillReport(10, "52.5"))

			assert.ErrorIs(t, reportError, errStorage)
		})
	}
}

func TestTradeOrderApplicationSkipsOrders(t *testing.T) {
	cases := []struct {
		name           string
		order          entities.TradeOrder
		expectedStatus string
		expectedFilled int
	}{
		{name: "skipping a waiting order", order: aWaitingBuyOrder(90), expectedStatus: "skipped", expectedFilled: 0},
		{name: "skipping a partly filled order keeps what filled", order: aWaitingBuyOrder(90, 40), expectedStatus: "skipped", expectedFilled: 40},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testHarness := newHarness(t, "2026-11-26 13:00")
			testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(testCase.order, true, nil)
			testHarness.orderRepository.EXPECT().MarkSkipped(gomock.Any(), "order-1").Return(nil)

			orderDto, skipError := testHarness.tradeOrderApplication.SkipOrder(t.Context(), "order-1")

			require.NoError(t, skipError)
			assert.Equal(t, testCase.expectedStatus, orderDto.Status)
			assert.Equal(t, testCase.expectedFilled, orderDto.FilledShares)
		})
	}
}

func TestTradeOrderApplicationRefusesToSkipAFilledOrder(t *testing.T) {
	testHarness := newHarness(t, "2026-11-26 13:00")
	testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), "order-1").Return(aWaitingBuyOrder(90, 90), true, nil)

	_, skipError := testHarness.tradeOrderApplication.SkipOrder(t.Context(), "order-1")

	assert.ErrorIs(t, skipError, domains.ErrConflict)
}

func TestTradeOrderApplicationPassesOnStorageFailuresWhileSkipping(t *testing.T) {
	t.Run("finding the order", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-26 13:00")
		testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(entities.TradeOrder{}, false, errStorage)

		_, skipError := testHarness.tradeOrderApplication.SkipOrder(t.Context(), "order-1")

		assert.ErrorIs(t, skipError, errStorage)
	})

	t.Run("recording the skip", func(t *testing.T) {
		testHarness := newHarness(t, "2026-11-26 13:00")
		testHarness.orderRepository.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(aWaitingBuyOrder(90), true, nil)
		testHarness.orderRepository.EXPECT().MarkSkipped(gomock.Any(), gomock.Any()).Return(errStorage)

		_, skipError := testHarness.tradeOrderApplication.SkipOrder(t.Context(), "order-1")

		assert.ErrorIs(t, skipError, errStorage)
	})
}
