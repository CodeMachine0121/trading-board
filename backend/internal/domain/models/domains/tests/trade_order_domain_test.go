package domains_test

import (
	"time"

	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func anOrderOf(side string, shares int, filledShares ...int) entities.TradeOrder {
	order := entities.TradeOrder{ID: "order-1", Side: side, Symbol: "2303", Shares: shares, TradeDate: "2026-11-26"}
	for _, filled := range filledShares {
		order.Fills = append(order.Fills, entities.OrderFill{Shares: filled})
	}

	return order
}

func aFillOf(shares int, price string) dto.OrderFillReportDto {
	return dto.OrderFillReportDto{Price: money(price), Shares: shares, FilledAt: taipeiTime("2026-11-26 09:05")}
}

func TestTradeOrderDomainProgressFollowsFilledShares(t *testing.T) {
	cases := []struct {
		name           string
		alreadyFilled  []int
		reportedShares int
		expectedStatus string
		expectedFilled int
		expectedFills  int
	}{
		{name: "a full fill completes the order", reportedShares: 90, expectedStatus: "filled", expectedFilled: 90, expectedFills: 1},
		{name: "a partial fill", reportedShares: 40, expectedStatus: "partiallyFilled", expectedFilled: 40, expectedFills: 1},
		{name: "the last shares complete the order", alreadyFilled: []int{40}, reportedShares: 50, expectedStatus: "filled", expectedFilled: 90, expectedFills: 2},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			orderDomain := domains.NewTradeOrderDomain(anOrderOf("buy", 90, testCase.alreadyFilled...), calendar)

			_, fillError := orderDomain.ReportFill(aFillOf(testCase.reportedShares, "52.5"),
				domains.NewFeeScheduleDomain(defaultFeeSchedule), taipeiTime("2026-11-26 13:00"))

			require.NoError(t, fillError)
			orderDto := orderDomain.ToDto()
			assert.Equal(t, testCase.expectedStatus, orderDto.Status)
			assert.Equal(t, testCase.expectedFilled, orderDto.FilledShares)
			assert.Len(t, orderDto.Fills, testCase.expectedFills)
		})
	}
}

func TestTradeOrderDomainRefusesFillsItCannotTake(t *testing.T) {
	skippedOrder := anOrderOf("buy", 90)
	skippedOrder.IsSkipped = true
	negativeFee := aFillOf(10, "52.5")
	negativeFee.Fee = decimal.NewNullDecimal(money("-1"))
	noTime := aFillOf(10, "52.5")
	noTime.FilledAt = time.Time{}

	cases := []struct {
		name            string
		order           entities.TradeOrder
		report          dto.OrderFillReportDto
		expectedKind    error
		expectedMessage string
	}{
		{name: "overfilling", order: anOrderOf("buy", 90, 40), report: aFillOf(51, "52.5"), expectedKind: domains.ErrValidation, expectedMessage: "超過這筆委託尚未成交的 50 股"},
		{name: "a skipped order", order: skippedOrder, report: aFillOf(10, "52.5"), expectedKind: domains.ErrConflict, expectedMessage: "已略過"},
		{name: "a zero price", order: anOrderOf("buy", 90), report: aFillOf(10, "0"), expectedKind: domains.ErrValidation},
		{name: "zero shares", order: anOrderOf("buy", 90), report: aFillOf(0, "52.5"), expectedKind: domains.ErrValidation},
		{name: "a negative fee", order: anOrderOf("buy", 90), report: negativeFee, expectedKind: domains.ErrValidation},
		{name: "no fill time", order: anOrderOf("buy", 90), report: noTime, expectedKind: domains.ErrValidation},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			orderDomain := domains.NewTradeOrderDomain(testCase.order, calendar)
			filledBefore := orderDomain.FilledShares()

			_, fillError := orderDomain.ReportFill(testCase.report,
				domains.NewFeeScheduleDomain(defaultFeeSchedule), taipeiTime("2026-11-26 13:00"))

			require.ErrorIs(t, fillError, testCase.expectedKind)
			assert.Contains(t, fillError.Error(), testCase.expectedMessage)
			assert.Equal(t, filledBefore, orderDomain.FilledShares())
		})
	}
}

func TestTradeOrderDomainWorksOutAmountAndFee(t *testing.T) {
	givenFee := aFillOf(90, "52.5")
	givenFee.Fee = decimal.NewNullDecimal(money("20"))

	cases := []struct {
		name           string
		side           string
		report         dto.OrderFillReportDto
		expectedAmount string
		expectedFee    string
	}{
		{name: "buy fee estimated when left out", side: "buy", report: aFillOf(90, "52.5"), expectedAmount: "4725", expectedFee: "4"},
		{name: "minimum fee on a tiny buy", side: "buy", report: aFillOf(1, "10"), expectedAmount: "10", expectedFee: "1"},
		{name: "sell fee includes transaction tax", side: "sell", report: aFillOf(50, "68"), expectedAmount: "3400", expectedFee: "13"},
		{name: "a given fee is kept", side: "buy", report: givenFee, expectedAmount: "4725", expectedFee: "20"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			orderDomain := domains.NewTradeOrderDomain(anOrderOf(testCase.side, 90), calendar)

			fill, fillError := orderDomain.ReportFill(testCase.report,
				domains.NewFeeScheduleDomain(defaultFeeSchedule), taipeiTime("2026-11-26 13:00"))

			require.NoError(t, fillError)
			assert.Equal(t, testCase.expectedAmount, fill.Amount.String())
			assert.Equal(t, testCase.expectedFee, fill.Fee.String())
			entry := orderDomain.LedgerEntryOfLatestFill()
			assert.Equal(t, testCase.expectedAmount, entry.Amount.String())
			assert.Equal(t, "2303", entry.Symbol)
		})
	}
}

func TestTradeOrderDomainSkipping(t *testing.T) {
	alreadySkipped := anOrderOf("buy", 90)
	alreadySkipped.IsSkipped = true

	cases := []struct {
		name           string
		order          entities.TradeOrder
		expectedError  error
		expectedStatus string
		expectedFilled int
	}{
		{name: "skipping a waiting order", order: anOrderOf("buy", 90), expectedStatus: "skipped", expectedFilled: 0},
		{name: "skipping a partly filled order keeps what filled", order: anOrderOf("buy", 90, 40), expectedStatus: "skipped", expectedFilled: 40},
		{name: "skipping an already skipped order changes nothing", order: alreadySkipped, expectedStatus: "skipped", expectedFilled: 0},
		{name: "a fully filled order cannot be skipped", order: anOrderOf("buy", 90, 90), expectedError: domains.ErrConflict, expectedStatus: "filled", expectedFilled: 90},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			orderDomain := domains.NewTradeOrderDomain(testCase.order, calendar)

			skipError := orderDomain.Skip()

			if testCase.expectedError != nil {
				require.ErrorIs(t, skipError, testCase.expectedError)
			} else {
				require.NoError(t, skipError)
			}
			assert.Equal(t, testCase.expectedStatus, orderDomain.ToDto().Status)
			assert.Equal(t, testCase.expectedFilled, orderDomain.ToDto().FilledShares)
		})
	}
}
