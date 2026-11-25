package application_test

import (
	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPortfolioApplicationListsHoldingsFromTheFillLog(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00").
		withFillLog("2026-11-20", aHeldBuyOrder("2303", 100, "50", "4", "2026-11-20 09:05"))

	positions, listError := testHarness.portfolioApplication.ListPositions(t.Context(), dto.PositionQueryDto{Status: "open"})

	require.NoError(t, listError)
	require.Len(t, positions, 1)
	assert.Equal(t, 100, positions[0].Shares)
	assert.Equal(t, "5004", positions[0].TotalCost.String())
	assert.Equal(t, "50.04", positions[0].AverageCost.String())
	assert.Equal(t, 5, positions[0].DaysHeld)
}

func TestPortfolioApplicationSummarisesAMonth(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00").
		withFillLog("2026-11-20", aHeldBuyOrder("2303", 90, "52.5", "4", "2026-11-20 09:05"))

	summary, summaryError := testHarness.portfolioApplication.GetMonthSummary(t.Context(), "2026-11")
	_, missingError := testHarness.portfolioApplication.GetMonthSummary(t.Context(), "2026-10")

	require.NoError(t, summaryError)
	assert.Equal(t, "4729", summary.Spent.String())
	assert.Equal(t, "5271", summary.Remaining.String())
	assert.Equal(t, "monthEnd", summary.Mode)
	assert.ErrorIs(t, missingError, domains.ErrNotFound)
}

func TestPortfolioApplicationPassesOnStorageFailures(t *testing.T) {
	testHarness := newHarness(t, "2026-11-25 20:00")
	testHarness.orderRepository.EXPECT().ListWithFills(gomock.Any()).Return(nil, errStorage).Times(2)

	_, listError := testHarness.portfolioApplication.ListPositions(t.Context(), dto.PositionQueryDto{})
	_, summaryError := testHarness.portfolioApplication.GetMonthSummary(t.Context(), "2026-11")

	assert.ErrorIs(t, listError, errStorage)
	assert.ErrorIs(t, summaryError, errStorage)
}
