package domains_test

import (
	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	buy  = vo.OrderSideBuy
	sell = vo.OrderSideSell
)

func openPositionsOf(t *testing.T, ledger domains.TradeLedgerDomain) []dto.PositionDto {
	positions, listError := ledger.Positions(dto.PositionQueryDto{})
	require.NoError(t, listError)

	return positions
}

func closedPositionsOf(t *testing.T, ledger domains.TradeLedgerDomain, month string) []dto.PositionDto {
	positions, listError := ledger.Positions(dto.PositionQueryDto{Status: "closed", Month: month})
	require.NoError(t, listError)

	return positions
}

func summaryOf(t *testing.T, ledger domains.TradeLedgerDomain, month string) dto.MonthSummaryDto {
	summary, summaryError := ledger.MonthSummary(month)
	require.NoError(t, summaryError)

	return summary
}

func TestTradeLedgerDomainHoldsAtAverageCost(t *testing.T) {
	cases := []struct {
		name                string
		book                *ledgerBook
		expectedShares      int
		expectedTotalCost   string
		expectedAverageCost string
		expectedRealized    string
	}{
		{
			name:           "first buy opens a holding",
			book:           (&ledgerBook{}).add(buy, "2303", "2026-11-20", 100, "50", "4", "2026-11-20 09:05"),
			expectedShares: 100, expectedTotalCost: "5004", expectedAverageCost: "50.04", expectedRealized: "0",
		},
		{
			name: "a second buy averages the cost",
			book: (&ledgerBook{}).
				add(buy, "2303", "2026-11-20", 100, "50", "4", "2026-11-20 09:05").
				add(buy, "2303", "2026-11-21", 100, "60", "5", "2026-11-21 09:05"),
			expectedShares: 200, expectedTotalCost: "11009", expectedAverageCost: "55.045", expectedRealized: "0",
		},
		{
			name: "a partial sale realizes profit at average cost",
			book: (&ledgerBook{}).
				add(buy, "2303", "2026-11-20", 100, "50", "4", "2026-11-20 09:05").
				add(buy, "2303", "2026-11-21", 100, "60", "5", "2026-11-21 09:05").
				add(sell, "2303", "2026-11-24", 100, "70", "17", "2026-11-24 09:05"),
			expectedShares: 100, expectedTotalCost: "5504.5", expectedAverageCost: "55.045", expectedRealized: "1478.5",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			positions := openPositionsOf(t, testCase.book.ledgerAt("2026-11-25 20:00"))

			require.Len(t, positions, 1)
			assert.Equal(t, testCase.expectedShares, positions[0].Shares)
			assert.Equal(t, testCase.expectedTotalCost, positions[0].TotalCost.String())
			assert.Equal(t, testCase.expectedAverageCost, positions[0].AverageCost.String())
			assert.Equal(t, testCase.expectedRealized, positions[0].RealizedPnl.String())
		})
	}
}

func TestTradeLedgerDomainClosesAHoldingAsAWinOrALoss(t *testing.T) {
	cases := []struct {
		name             string
		book             *ledgerBook
		expectedRealized string
		expectedResult   string
	}{
		{
			name: "selling the rest closes the holding as a win",
			book: (&ledgerBook{}).
				add(buy, "2303", "2026-11-20", 100, "50", "4", "2026-11-20 09:05").
				add(buy, "2303", "2026-11-21", 100, "60", "5", "2026-11-21 09:05").
				add(sell, "2303", "2026-11-24", 100, "70", "17", "2026-11-24 09:05").
				add(sell, "2303", "2026-11-25", 100, "50", "12", "2026-11-25 09:05"),
			expectedRealized: "962", expectedResult: "win",
		},
		{
			name: "a closed holding with no profit is a loss",
			book: (&ledgerBook{}).
				add(buy, "2303", "2026-11-20", 10, "10", "0", "2026-11-20 09:05").
				add(sell, "2303", "2026-11-25", 10, "10", "0", "2026-11-25 09:05"),
			expectedRealized: "0", expectedResult: "loss",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			ledger := testCase.book.ledgerAt("2026-11-25 20:00")

			closed := closedPositionsOf(t, ledger, "")

			assert.Empty(t, openPositionsOf(t, ledger))
			require.Len(t, closed, 1)
			assert.Equal(t, testCase.expectedRealized, closed[0].RealizedPnl.String())
			assert.Equal(t, testCase.expectedResult, *closed[0].Result)
			assert.Equal(t, "2026-11-25", *closed[0].ClosedDate)
			assert.Equal(t, 5, closed[0].DaysHeld)
		})
	}
}

func TestTradeLedgerDomainRealizesEachSaleSeparately(t *testing.T) {
	ledger := (&ledgerBook{}).
		add(buy, "2303", "2026-11-20", 100, "50", "4", "2026-11-20 09:05").
		add(buy, "2303", "2026-11-21", 100, "60", "5", "2026-11-21 09:05").
		add(sell, "2303", "2026-11-24", 100, "70", "17", "2026-11-24 09:05").
		add(sell, "2303", "2026-12-01", 100, "50", "12", "2026-12-01 09:05").
		ledgerAt("2026-12-02 20:00")

	assert.Equal(t, "1478.5", summaryOf(t, ledger, "2026-11").RealizedPnl.String())
	assert.Equal(t, "-516.5", summaryOf(t, ledger, "2026-12").RealizedPnl.String())
}

func TestTradeLedgerDomainOpensANewHoldingAfterClosing(t *testing.T) {
	ledger := (&ledgerBook{}).
		add(buy, "2303", "2026-11-20", 100, "50", "4", "2026-11-20 09:05").
		add(sell, "2303", "2026-11-23", 100, "60", "10", "2026-11-23 09:05").
		add(buy, "2303", "2026-11-24", 10, "55", "1", "2026-11-24 09:05").
		ledgerAt("2026-11-25 20:00")

	open := openPositionsOf(t, ledger)
	closed := closedPositionsOf(t, ledger, "2026-11")

	require.Len(t, open, 1)
	assert.Equal(t, 10, open[0].Shares)
	assert.Equal(t, "2026-11-24", open[0].OpenedDate)
	require.Len(t, closed, 1)
	assert.Equal(t, "2026-11-20", closed[0].OpenedDate)
	assert.Equal(t, 100, closed[0].Shares)
	assert.Equal(t, "986", closed[0].RealizedPnl.String()) // 6000 - 10 - 5004
}

func TestTradeLedgerDomainShowsOpenHoldingDetails(t *testing.T) {
	ledger := (&ledgerBook{}).
		add(buy, "2303", "2026-11-20", 100, "50", "4", "2026-11-20 09:05").
		ledgerAt("2026-11-25 20:00")

	position := openPositionsOf(t, ledger)[0]

	assert.Equal(t, "2303", position.Symbol)
	assert.Equal(t, "open", position.Status)
	assert.Equal(t, "48", position.StopLossPrice.String())
	assert.Equal(t, "60", position.TakeProfitPrice.String())
	assert.Equal(t, "2026-11-20", position.OpenedDate)
	assert.Equal(t, 5, position.DaysHeld)
	assert.Nil(t, position.ClosedDate)
	assert.Nil(t, position.Result)
}

func TestTradeLedgerDomainTakesStopAndTakeProfitFromTheLatestBuyOrder(t *testing.T) {
	book := &ledgerBook{}
	book.add(buy, "2303", "2026-11-21", 100, "50", "4", "2026-11-21 09:05")
	olderOrderFilledLater := book.entryOf(buy, "2303", "2026-11-20", 10, "50", "1", "2026-11-22 09:05")
	olderOrderFilledLater.StopLossPrice = decimal.NewNullDecimal(money("40"))
	book.entries = append(book.entries, olderOrderFilledLater)

	position := openPositionsOf(t, book.ledgerAt("2026-11-25 20:00"))[0]

	assert.Equal(t, "48", position.StopLossPrice.String())
	assert.Equal(t, 110, position.Shares)
}

func TestTradeLedgerDomainListsClosedHoldingsOfAMonth(t *testing.T) {
	ledger := (&ledgerBook{}).
		add(buy, "2303", "2026-11-10", 10, "50", "1", "2026-11-10 09:05").
		add(buy, "3006", "2026-11-11", 10, "60", "1", "2026-11-11 09:05").
		add(buy, "2330", "2026-11-12", 10, "70", "1", "2026-11-12 09:05").
		add(sell, "2303", "2026-11-24", 10, "55", "1", "2026-11-24 09:05").
		add(sell, "3006", "2026-12-02", 10, "65", "1", "2026-12-02 09:05").
		add(sell, "2330", "2026-11-25", 10, "65", "1", "2026-11-25 09:05").
		ledgerAt("2026-12-03 20:00")

	november := closedPositionsOf(t, ledger, "2026-11")
	every := closedPositionsOf(t, ledger, "")

	require.Len(t, november, 2)
	assert.Equal(t, []string{"2330", "2303"}, []string{november[0].Symbol, november[1].Symbol})
	require.Len(t, every, 3)
	assert.Equal(t, "3006", every[0].Symbol)
}

func TestTradeLedgerDomainRefusesUnreadablePositionQueries(t *testing.T) {
	ledger := (&ledgerBook{}).ledgerAt("2026-11-25 20:00")

	_, statusError := ledger.Positions(dto.PositionQueryDto{Status: "sold"})
	_, monthError := ledger.Positions(dto.PositionQueryDto{Status: "closed", Month: "2026/11"})

	assert.ErrorIs(t, statusError, domains.ErrValidation)
	assert.ErrorIs(t, monthError, domains.ErrValidation)
}

func TestTradeLedgerDomainIgnoresASaleWithNothingHeld(t *testing.T) {
	ledger := (&ledgerBook{}).
		add(sell, "2303", "2026-11-20", 10, "50", "1", "2026-11-20 09:05").
		ledgerAt("2026-11-25 20:00")

	assert.Empty(t, closedPositionsOf(t, ledger, ""))
	assert.Equal(t, 0, ledger.HeldShares("2303"))
}

func TestTradeLedgerDomainRollsTheBudgetOver(t *testing.T) {
	book := (&ledgerBook{}).
		add(buy, "2303", "2026-11-20", 90, "52.5", "4", "2026-11-20 09:05")

	november := summaryOf(t, book.ledgerAt("2026-12-05 20:00"), "2026-11")
	december := summaryOf(t, book.ledgerAt("2026-12-05 20:00"), "2026-12")

	assert.Equal(t, "10000", november.Allocation.String())
	assert.Equal(t, "0", november.CarriedOver.String())
	assert.Equal(t, "4729", november.Spent.String())
	assert.Equal(t, "5271", november.Remaining.String())
	assert.Equal(t, "5271", december.CarriedOver.String())
	assert.Equal(t, "15271", december.Remaining.String())
}

func TestTradeLedgerDomainRecyclesSaleProceeds(t *testing.T) {
	ledger := (&ledgerBook{}).
		add(buy, "3006", "2026-11-20", 50, "60", "2", "2026-11-20 09:05").
		add(sell, "3006", "2026-12-03", 50, "68", "13", "2026-12-03 09:05").
		ledgerAt("2026-12-05 20:00")

	assert.Equal(t, "3387", summaryOf(t, ledger, "2026-12").RecycledProceeds.String())
}

func TestTradeLedgerDomainAdmitsPurchasesOnlyWithinTheBudget(t *testing.T) {
	cases := []struct {
		name              string
		firstBuyPrice     string
		isAdmitted        bool
		expectedRemaining string
	}{
		{name: "a buy that uses exactly the remaining money is accepted", firstBuyPrice: "5270", isAdmitted: true, expectedRemaining: "0"},
		{name: "a buy that would overspend is refused", firstBuyPrice: "5271", isAdmitted: false, expectedRemaining: "4728"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			book := (&ledgerBook{}).add(buy, "2330", "2026-11-20", 1, testCase.firstBuyPrice, "1", "2026-11-20 09:05")
			ledger := book.ledgerAt("2026-11-25 20:00")
			purchase := book.entryOf(buy, "2303", "2026-11-25", 90, "52.5", "4", "2026-11-25 09:05")

			admitError := ledger.Admit(purchase)

			if testCase.isAdmitted {
				require.NoError(t, admitError)
				book.entries = append(book.entries, purchase)
			} else {
				require.ErrorIs(t, admitError, domains.ErrConflict)
				assert.Contains(t, admitError.Error(), "預算不足")
			}
			assert.Equal(t, testCase.expectedRemaining,
				summaryOf(t, book.ledgerAt("2026-11-25 20:00"), "2026-11").Remaining.String())
		})
	}
}

func TestTradeLedgerDomainRefusesABackdatedPurchaseThatALaterMonthCannotAfford(t *testing.T) {
	book := (&ledgerBook{}).
		add(buy, "2330", "2026-11-03", 1, "1", "0", "2026-11-03 09:05").
		add(buy, "2303", "2026-12-03", 1, "15000", "0", "2026-12-03 09:05")
	backdated := book.entryOf(buy, "3006", "2026-11-04", 1, "6000", "0", "2026-11-04 09:05")

	admitError := book.ledgerAt("2026-12-05 20:00").Admit(backdated)

	require.ErrorIs(t, admitError, domains.ErrConflict)
	assert.Contains(t, admitError.Error(), "2026-12")
}

func TestTradeLedgerDomainAdmitsSalesOnlyFromTheHolding(t *testing.T) {
	book := (&ledgerBook{}).add(buy, "3006", "2026-11-20", 50, "60", "2", "2026-11-20 09:05")
	ledger := book.ledgerAt("2026-11-25 20:00")

	tooMany := ledger.Admit(book.entryOf(sell, "3006", "2026-11-25", 60, "68", "13", "2026-11-25 09:05"))
	allOfIt := ledger.Admit(book.entryOf(sell, "3006", "2026-11-25", 50, "68", "13", "2026-11-25 09:05"))

	require.ErrorIs(t, tooMany, domains.ErrConflict)
	assert.Contains(t, tooMany.Error(), "只持有 50 股")
	assert.NoError(t, allOfIt)
}

func TestTradeLedgerDomainRequiresHoldingsForPublishedSales(t *testing.T) {
	ledger := (&ledgerBook{}).add(buy, "3006", "2026-11-20", 50, "60", "2", "2026-11-20 09:05").ledgerAt("2026-11-25 20:00")

	cases := []struct {
		name       string
		sales      map[string]int
		isAccepted bool
	}{
		{name: "selling what is held", sales: map[string]int{"3006": 50}, isAccepted: true},
		{name: "selling one share more than held", sales: map[string]int{"3006": 51}, isAccepted: false},
		{name: "selling a stock not held", sales: map[string]int{"2330": 10}, isAccepted: false},
		{name: "two sells together exceeding the holding", sales: map[string]int{"3006": 60}, isAccepted: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			holdingError := ledger.RequireHoldings(testCase.sales)

			if testCase.isAccepted {
				assert.NoError(t, holdingError)
				return
			}
			require.ErrorIs(t, holdingError, domains.ErrValidation)
			assert.Contains(t, holdingError.Error(), "超過持股")
		})
	}
}

func TestTradeLedgerDomainKnowsWhichMonthsExist(t *testing.T) {
	book := (&ledgerBook{}).add(buy, "2303", "2026-11-20", 1, "10", "1", "2026-11-20 09:05")

	_, beforeFirst := book.ledgerAt("2026-11-25 20:00").MonthSummary("2026-10")
	_, unreadable := book.ledgerAt("2026-11-25 20:00").MonthSummary("2026-1")
	cardOnly := domains.NewTradeLedgerDomain(nil, "2026-09-30", calendar, defaultBudgetPolicy, taipeiTime("2026-11-25 20:00"))
	nothingYet := domains.NewTradeLedgerDomain(nil, "", calendar, defaultBudgetPolicy, taipeiTime("2026-11-25 20:00"))

	require.ErrorIs(t, beforeFirst, domains.ErrNotFound)
	require.ErrorIs(t, unreadable, domains.ErrValidation)
	assert.Equal(t, "20000", summaryOf(t, cardOnly, "2026-10").Remaining.String())
	assert.Equal(t, "10000", summaryOf(t, nothingYet, "2026-11").Remaining.String())
	_, beforeNothing := nothingYet.MonthSummary("2026-10")
	assert.ErrorIs(t, beforeNothing, domains.ErrNotFound)
}

func TestTradeLedgerDomainReportsTheBudgetMode(t *testing.T) {
	cases := []struct {
		name         string
		now          string
		month        string
		policy       vo.BudgetPolicyVo
		expectedMode string
	}{
		{name: "strict on the 21st", now: "2026-11-21 20:00", month: "2026-11", policy: defaultBudgetPolicy, expectedMode: "strict"},
		{name: "month-end from the 22nd", now: "2026-11-22 08:00", month: "2026-11", policy: defaultBudgetPolicy, expectedMode: "monthEnd"},
		{name: "a past month is month-end", now: "2026-12-02 08:00", month: "2026-11", policy: defaultBudgetPolicy, expectedMode: "monthEnd"},
		{name: "a future month is strict", now: "2026-11-25 08:00", month: "2026-12", policy: defaultBudgetPolicy, expectedMode: "strict"},
		{name: "an impossible start day falls back to the 22nd", now: "2026-11-21 20:00", month: "2026-11",
			policy: vo.BudgetPolicyVo{MonthlyAllocation: money("10000"), MonthEndStartDay: 0}, expectedMode: "strict"},
		{name: "a configured start day of the 15th", now: "2026-11-15 08:00", month: "2026-11",
			policy: vo.BudgetPolicyVo{MonthlyAllocation: money("10000"), MonthEndStartDay: 15}, expectedMode: "monthEnd"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			ledger := domains.NewTradeLedgerDomain(nil, "2026-11-01", calendar, testCase.policy, taipeiTime(testCase.now))

			assert.Equal(t, testCase.expectedMode, summaryOf(t, ledger, testCase.month).Mode)
		})
	}
}

func TestTradeLedgerDomainTreatsANegativeAllocationAsZero(t *testing.T) {
	ledger := domains.NewTradeLedgerDomain(nil, "2026-11-01", calendar,
		vo.BudgetPolicyVo{MonthlyAllocation: money("-5"), MonthEndStartDay: 22}, taipeiTime("2026-11-25 20:00"))

	assert.Equal(t, "0", summaryOf(t, ledger, "2026-11").Allocation.String())
}

func TestTradeLedgerDomainReportsMonthlyPerformance(t *testing.T) {
	cases := []struct {
		name             string
		book             *ledgerBook
		expectedRealized string
		expectedClosed   int
		expectedWins     int
		expectedLosses   int
		expectedWinRate  string
		expectedReturn   string
	}{
		{
			name: "a win of 962 and a loss of 300 on 11009 of cost",
			book: (&ledgerBook{}).
				add(buy, "2303", "2026-11-02", 100, "50", "4", "2026-11-02 09:05").
				add(sell, "2303", "2026-11-05", 100, "59.66", "0", "2026-11-05 09:05").
				add(buy, "3006", "2026-11-06", 100, "60.05", "0", "2026-11-06 09:05").
				add(sell, "3006", "2026-11-10", 100, "57.05", "0", "2026-11-10 09:05"),
			expectedRealized: "662", expectedClosed: 2, expectedWins: 1, expectedLosses: 1,
			expectedWinRate: "50", expectedReturn: "6.01",
		},
		{
			name:             "a month with nothing sold",
			book:             (&ledgerBook{}).add(buy, "2303", "2026-11-02", 100, "50", "4", "2026-11-02 09:05"),
			expectedRealized: "0", expectedClosed: 0, expectedWins: 0, expectedLosses: 0,
			expectedWinRate: "0", expectedReturn: "0",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			summary := summaryOf(t, testCase.book.ledgerAt("2026-11-25 20:00"), "2026-11")

			assert.Equal(t, testCase.expectedRealized, summary.RealizedPnl.String())
			assert.Equal(t, testCase.expectedClosed, summary.ClosedPositionCount)
			assert.Equal(t, testCase.expectedWins, summary.WinCount)
			assert.Equal(t, testCase.expectedLosses, summary.LossCount)
			assert.Equal(t, testCase.expectedWinRate, summary.WinRate.String())
			assert.Equal(t, testCase.expectedReturn, summary.RealizedReturnPct.String())
		})
	}
}

func TestTradeLedgerDomainNeverReportsNegativeDaysHeld(t *testing.T) {
	ledger := (&ledgerBook{}).add(buy, "2303", "2026-11-27", 1, "10", "1", "2026-11-27 09:05").ledgerAt("2026-11-25 20:00")

	assert.Equal(t, 0, openPositionsOf(t, ledger)[0].DaysHeld)
}
