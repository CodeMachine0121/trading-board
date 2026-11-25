package controller_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/middlewares"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/router"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/interface/mocks"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	agentKey   = "agent-key"
	ownerToken = "owner-pass"
)

var taipei, _ = time.LoadLocation("Asia/Taipei")

// api serves the real route table over the real applications, services and domain models; only the
// outermost interfaces (storage, clock, pass signing) are mocked.
type api struct {
	engine           *gin.Engine
	cardRepository   *mocks.MockITradeCardRepository
	orderRepository  *mocks.MockITradeOrderRepository
	fillRepository   *mocks.MockIOrderFillRepository
	accessTokenProxy *mocks.MockIAccessTokenProxy
}

func newApi(t *testing.T) *api {
	gin.SetMode(gin.TestMode)
	mockController := gomock.NewController(t)
	testApi := &api{
		engine:           gin.New(),
		cardRepository:   mocks.NewMockITradeCardRepository(mockController),
		orderRepository:  mocks.NewMockITradeOrderRepository(mockController),
		fillRepository:   mocks.NewMockIOrderFillRepository(mockController),
		accessTokenProxy: mocks.NewMockIAccessTokenProxy(mockController),
	}
	transactionRepository := mocks.NewMockITransactionRepository(mockController)
	transactionRepository.EXPECT().Exclusively(gomock.Any(), gomock.Any()).DoAndReturn(
		func(executionContext context.Context, work func(context.Context) error) error {
			return work(executionContext)
		}).AnyTimes()
	clockProxy := mocks.NewMockIClockProxy(mockController)
	clockProxy.EXPECT().Now().Return(time.Date(2026, 11, 25, 20, 0, 0, 0, taipei)).AnyTimes()
	testApi.accessTokenProxy.EXPECT().Verify(ownerToken).Return(vo.AccessTokenClaimsVo{Subject: "owner"}, nil).AnyTimes()
	testApi.accessTokenProxy.EXPECT().Verify(gomock.Not(ownerToken)).Return(vo.AccessTokenClaimsVo{}, domains.AuthenticationRequired("bad")).AnyTimes()

	calendar := domains.NewBusinessCalendarDomain(taipei)
	ledgerLoader := service.NewTradeLedgerLoader(testApi.orderRepository, testApi.cardRepository, clockProxy, calendar,
		vo.BudgetPolicyVo{MonthlyAllocation: decimal.NewFromInt(10000), MonthEndStartDay: 22})
	authenticationApplication := application.NewAuthenticationApplication(service.NewAuthenticationService(
		domains.NewAccessCredentialDomain(vo.AccessCredentialVo{OwnerUsername: "james", OwnerPassword: "secret", AgentApiKey: agentKey}),
		24*time.Hour, testApi.accessTokenProxy, clockProxy, calendar))
	errorPresenter := controller.NewErrorPresenter()

	router.NewRouter(
		middlewares.NewCorsMiddleware([]string{"https://trading-board.coding-afternoon.com"}),
		middlewares.NewAuthenticationMiddleware(authenticationApplication, errorPresenter),
		controller.NewAuthenticationController(authenticationApplication, errorPresenter),
		controller.NewTradeCardController(application.NewTradeCardApplication(service.NewTradeCardService(
			testApi.cardRepository, transactionRepository, ledgerLoader, clockProxy, calendar)), errorPresenter),
		controller.NewTradeOrderController(application.NewTradeOrderApplication(service.NewTradeOrderService(
			testApi.orderRepository, testApi.fillRepository, transactionRepository, ledgerLoader, clockProxy, calendar,
			domains.NewFeeScheduleDomain(vo.FeeScheduleVo{BrokerFeeRate: decimal.RequireFromString("0.001425"),
				Discount: decimal.RequireFromString("0.6"), MinimumFee: decimal.NewFromInt(1),
				TransactionTaxRate: decimal.RequireFromString("0.003")}))), errorPresenter),
		controller.NewPortfolioController(application.NewPortfolioApplication(service.NewPortfolioService(ledgerLoader)), errorPresenter),
		controller.NewHealthController(),
	).Register(testApi.engine)

	return testApi
}

func (testApi *api) withEmptyFillLog() *api {
	testApi.orderRepository.EXPECT().ListWithFills(gomock.Any()).Return(nil, nil).AnyTimes()
	testApi.cardRepository.EXPECT().FindEarliestTradeDate(gomock.Any()).Return("", false, nil).AnyTimes()

	return testApi
}

type response struct {
	status int
	body   map[string]any
	list   []map[string]any
	header http.Header
}

func (testApi *api) call(t *testing.T, method string, path string, body string, headers map[string]string) response {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	recorder := httptest.NewRecorder()
	testApi.engine.ServeHTTP(recorder, request)

	raw, readError := io.ReadAll(recorder.Body)
	require.NoError(t, readError)
	result := response{status: recorder.Code, header: recorder.Header()}
	if len(raw) > 0 && raw[0] == '[' {
		require.NoError(t, json.Unmarshal(raw, &result.list))
	} else if len(raw) > 0 {
		require.NoError(t, json.Unmarshal(raw, &result.body))
	}

	return result
}

func asAgent() map[string]string {
	return map[string]string{"X-Api-Key": agentKey}
}

func asOwner() map[string]string {
	return map[string]string{"Authorization": "Bearer " + ownerToken}
}

func errorCodeOf(result response) string {
	errorBody, _ := result.body["error"].(map[string]any)
	code, _ := errorBody["code"].(string)

	return code
}

func storedCard() entities.TradeCard {
	return entities.TradeCard{ID: "card-1", TradeDate: "2026-11-26", Status: "trade", Note: "note",
		PublishedAt: time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC),
		Orders: []entities.TradeOrder{{ID: "order-1", Side: "buy", Symbol: "2303", Name: "聯電", Shares: 90,
			LimitPrice: decimal.RequireFromString("52.5"), MaxChasePrice: decimal.NewNullDecimal(decimal.RequireFromString("54")),
			EstimatedAmount: decimal.NewNullDecimal(decimal.RequireFromString("4725")),
			StopLossPrice:   decimal.NewNullDecimal(decimal.RequireFromString("48")),
			TakeProfitPrice: decimal.NewNullDecimal(decimal.RequireFromString("60")), TradeDate: "2026-11-26"}}}
}
