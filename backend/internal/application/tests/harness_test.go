package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/interface/mocks"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/service"
	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

var taipei, _ = time.LoadLocation("Asia/Taipei")

func taipeiTime(text string) time.Time {
	parsed, parseError := time.ParseInLocation("2006-01-02 15:04", text, taipei)
	if parseError != nil {
		panic(parseError)
	}

	return parsed
}

func money(text string) decimal.Decimal {
	return decimal.RequireFromString(text)
}

// harness wires the real applications, services and domain models; only the outermost interfaces are mocked.
type harness struct {
	cardRepository        *mocks.MockITradeCardRepository
	orderRepository       *mocks.MockITradeOrderRepository
	fillRepository        *mocks.MockIOrderFillRepository
	transactionRepository *mocks.MockITransactionRepository
	clockProxy            *mocks.MockIClockProxy
	accessTokenProxy      *mocks.MockIAccessTokenProxy

	authenticationApplication *application.AuthenticationApplication
	tradeCardApplication      *application.TradeCardApplication
	tradeOrderApplication     *application.TradeOrderApplication
	portfolioApplication      *application.PortfolioApplication
}

func newHarness(t *testing.T, now string) *harness {
	controller := gomock.NewController(t)
	testHarness := &harness{
		cardRepository:        mocks.NewMockITradeCardRepository(controller),
		orderRepository:       mocks.NewMockITradeOrderRepository(controller),
		fillRepository:        mocks.NewMockIOrderFillRepository(controller),
		transactionRepository: mocks.NewMockITransactionRepository(controller),
		clockProxy:            mocks.NewMockIClockProxy(controller),
		accessTokenProxy:      mocks.NewMockIAccessTokenProxy(controller),
	}
	testHarness.clockProxy.EXPECT().Now().Return(taipeiTime(now)).AnyTimes()
	testHarness.transactionRepository.EXPECT().Exclusively(gomock.Any(), gomock.Any()).DoAndReturn(
		func(executionContext context.Context, work func(context.Context) error) error {
			return work(executionContext)
		}).AnyTimes()

	calendar := domains.NewBusinessCalendarDomain(taipei)
	ledgerLoader := service.NewTradeLedgerLoader(testHarness.orderRepository, testHarness.cardRepository,
		testHarness.clockProxy, calendar, vo.BudgetPolicyVo{MonthlyAllocation: money("10000"), MonthEndStartDay: 22})

	testHarness.authenticationApplication = application.NewAuthenticationApplication(service.NewAuthenticationService(
		domains.NewAccessCredentialDomain(vo.AccessCredentialVo{OwnerUsername: "james", OwnerPassword: "secret", AgentApiKey: "agentkey"}),
		24*time.Hour, testHarness.accessTokenProxy, testHarness.clockProxy, calendar))
	testHarness.tradeCardApplication = application.NewTradeCardApplication(service.NewTradeCardService(
		testHarness.cardRepository, testHarness.transactionRepository, ledgerLoader, testHarness.clockProxy, calendar))
	testHarness.tradeOrderApplication = application.NewTradeOrderApplication(service.NewTradeOrderService(
		testHarness.orderRepository, testHarness.fillRepository, testHarness.transactionRepository, ledgerLoader,
		testHarness.clockProxy, calendar, domains.NewFeeScheduleDomain(vo.FeeScheduleVo{
			BrokerFeeRate: money("0.001425"), Discount: money("0.6"), MinimumFee: money("1"), TransactionTaxRate: money("0.003"),
		})))
	testHarness.portfolioApplication = application.NewPortfolioApplication(service.NewPortfolioService(ledgerLoader))

	return testHarness
}

// withFillLog makes the ledger see these orders (with their fills) and an earliest card date.
func (testHarness *harness) withFillLog(earliestCardDate string, orders ...entities.TradeOrder) *harness {
	testHarness.orderRepository.EXPECT().ListWithFills(gomock.Any()).Return(orders, nil).AnyTimes()
	testHarness.cardRepository.EXPECT().FindEarliestTradeDate(gomock.Any()).
		Return(earliestCardDate, earliestCardDate != "", nil).AnyTimes()

	return testHarness
}

func aHeldBuyOrder(symbol string, shares int, price string, fee string, filledAt string) entities.TradeOrder {
	return entities.TradeOrder{
		ID: "held-" + symbol, Side: "buy", Symbol: symbol, Name: "name-" + symbol, TradeDate: filledAt[:10], Shares: shares,
		StopLossPrice: decimal.NewNullDecimal(money("1")), TakeProfitPrice: decimal.NewNullDecimal(money("1000")),
		Fills: []entities.OrderFill{{ID: "fill-" + symbol, TradeOrderID: "held-" + symbol, Shares: shares, Price: money(price),
			Amount: money(price).Mul(decimal.NewFromInt(int64(shares))), Fee: money(fee),
			FilledAt: taipeiTime(filledAt), RecordedAt: taipeiTime(filledAt)}},
	}
}

func gomockAny() gomock.Matcher {
	return gomock.Any()
}
