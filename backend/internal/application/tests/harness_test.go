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

// Fake credentials for tests only; the hash is never checked for real because the password proxy is mocked.
const (
	ownerEmail        = "owner@example.com"
	ownerPasswordHash = "$2a$10$test-only-hash"
	clientID          = "trading-board-web"
	redirectUri       = "http://localhost:3000/auth/callback"
)

func taipeiTimeWithSeconds(text string) time.Time {
	parsed, parseError := time.ParseInLocation("2006-01-02 15:04:05", text, taipei)
	if parseError != nil {
		panic(parseError)
	}

	return parsed
}

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
	apiKeyRepository      *mocks.MockIApiKeyRepository
	codeRepository        *mocks.MockIAuthorizationCodeRepository
	refreshRepository     *mocks.MockIRefreshTokenRepository
	failureRepository     *mocks.MockISignInFailureRepository
	passwordHashProxy     *mocks.MockIPasswordHashProxy
	randomTokenProxy      *mocks.MockIRandomTokenProxy

	// rolledBack counts units of work whose writes the real transaction would have rolled back (work returned an error).
	rolledBack int

	authenticationApplication *application.AuthenticationApplication
	authorizationApplication  *application.AuthorizationApplication
	apiKeyApplication         *application.ApiKeyApplication
	tradeCardApplication      *application.TradeCardApplication
	tradeOrderApplication     *application.TradeOrderApplication
	portfolioApplication      *application.PortfolioApplication
}

func newHarness(t *testing.T, now string) *harness {
	return newHarnessAt(t, taipeiTime(now))
}

func newHarnessAt(t *testing.T, now time.Time) *harness {
	controller := gomock.NewController(t)
	testHarness := &harness{
		cardRepository:        mocks.NewMockITradeCardRepository(controller),
		orderRepository:       mocks.NewMockITradeOrderRepository(controller),
		fillRepository:        mocks.NewMockIOrderFillRepository(controller),
		transactionRepository: mocks.NewMockITransactionRepository(controller),
		clockProxy:            mocks.NewMockIClockProxy(controller),
		accessTokenProxy:      mocks.NewMockIAccessTokenProxy(controller),
		apiKeyRepository:      mocks.NewMockIApiKeyRepository(controller),
		codeRepository:        mocks.NewMockIAuthorizationCodeRepository(controller),
		refreshRepository:     mocks.NewMockIRefreshTokenRepository(controller),
		failureRepository:     mocks.NewMockISignInFailureRepository(controller),
		passwordHashProxy:     mocks.NewMockIPasswordHashProxy(controller),
		randomTokenProxy:      mocks.NewMockIRandomTokenProxy(controller),
	}
	testHarness.clockProxy.EXPECT().Now().Return(now).AnyTimes()
	testHarness.transactionRepository.EXPECT().Exclusively(gomock.Any(), gomock.Any()).DoAndReturn(
		func(executionContext context.Context, work func(context.Context) error) error {
			workError := work(executionContext)
			if workError != nil {
				testHarness.rolledBack++
			}

			return workError
		}).AnyTimes()

	calendar := domains.NewBusinessCalendarDomain(taipei)
	ledgerLoader := service.NewTradeLedgerLoader(testHarness.orderRepository, testHarness.cardRepository,
		testHarness.clockProxy, calendar, vo.BudgetPolicyVo{MonthlyAllocation: money("10000"), MonthEndStartDay: 22})

	ownerAccount := domains.NewOwnerAccountDomain(vo.OwnerAccountVo{Email: ownerEmail, PasswordHash: ownerPasswordHash})
	testHarness.authenticationApplication = application.NewAuthenticationApplication(service.NewAuthenticationService(
		ownerAccount, testHarness.apiKeyRepository, testHarness.accessTokenProxy, testHarness.clockProxy, calendar))
	testHarness.authorizationApplication = application.NewAuthorizationApplication(service.NewAuthorizationService(
		ownerAccount, vo.OAuthClientVo{ClientID: clientID, RedirectUris: []string{redirectUri}},
		vo.AuthorizationPolicyVo{AuthorizationCodeLifetime: time.Minute, AccessTokenLifetime: 15 * time.Minute,
			RefreshTokenLifetime: 30 * 24 * time.Hour},
		testHarness.codeRepository, testHarness.refreshRepository, testHarness.failureRepository,
		testHarness.transactionRepository, testHarness.passwordHashProxy, testHarness.randomTokenProxy,
		testHarness.accessTokenProxy, testHarness.clockProxy))
	testHarness.apiKeyApplication = application.NewApiKeyApplication(service.NewApiKeyService(
		testHarness.apiKeyRepository, testHarness.transactionRepository, testHarness.randomTokenProxy,
		testHarness.clockProxy, calendar))
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

// withOwnerAccount rebuilds the sign-in over another configured owner account (for example none at all).
func (testHarness *harness) withOwnerAccount(account vo.OwnerAccountVo) *harness {
	testHarness.authorizationApplication = application.NewAuthorizationApplication(service.NewAuthorizationService(
		domains.NewOwnerAccountDomain(account), vo.OAuthClientVo{ClientID: clientID, RedirectUris: []string{redirectUri}},
		vo.AuthorizationPolicyVo{AuthorizationCodeLifetime: time.Minute, AccessTokenLifetime: 15 * time.Minute,
			RefreshTokenLifetime: 30 * 24 * time.Hour},
		testHarness.codeRepository, testHarness.refreshRepository, testHarness.failureRepository,
		testHarness.transactionRepository, testHarness.passwordHashProxy, testHarness.randomTokenProxy,
		testHarness.accessTokenProxy, testHarness.clockProxy))

	return testHarness
}
