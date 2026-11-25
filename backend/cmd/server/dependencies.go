package main

import (
	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/config"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/middlewares"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/router"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/service"
	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/clock"
	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/persistence"
	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/security"
	"gorm.io/gorm"
)

// newRouter is the composition root: the only place that knows every concrete type.
func newRouter(database *gorm.DB, applicationConfig config.ApplicationConfig) *router.Router {
	tradeCardRepository := persistence.NewTradeCardRepository(database)
	tradeOrderRepository := persistence.NewTradeOrderRepository(database)
	orderFillRepository := persistence.NewOrderFillRepository(database)
	transactionRepository := persistence.NewTransactionRepository(database)
	apiKeyRepository := persistence.NewApiKeyRepository(database)
	clockProxy := clock.NewSystemClockProxy()
	accessTokenProxy := security.NewJwtAccessTokenProxy(applicationConfig.AccessTokenSigningKey)
	randomTokenProxy := security.NewCryptoRandomTokenProxy()
	calendar := domains.NewBusinessCalendarDomain(applicationConfig.BusinessTimeZone)
	ownerAccount := domains.NewOwnerAccountDomain(applicationConfig.OwnerAccount)

	ledgerLoader := service.NewTradeLedgerLoader(
		tradeOrderRepository, tradeCardRepository, clockProxy, calendar, applicationConfig.BudgetPolicy)
	authenticationApplication := application.NewAuthenticationApplication(service.NewAuthenticationService(
		ownerAccount, apiKeyRepository, accessTokenProxy, clockProxy, calendar))
	authorizationApplication := application.NewAuthorizationApplication(service.NewAuthorizationService(
		ownerAccount, applicationConfig.OAuthClient, applicationConfig.AuthorizationPolicy,
		persistence.NewAuthorizationCodeRepository(database), persistence.NewRefreshTokenRepository(database),
		persistence.NewSignInFailureRepository(database), transactionRepository,
		security.NewBcryptPasswordHashProxy(), randomTokenProxy, accessTokenProxy, clockProxy))
	apiKeyApplication := application.NewApiKeyApplication(service.NewApiKeyService(
		apiKeyRepository, transactionRepository, randomTokenProxy, clockProxy, calendar))
	tradeCardApplication := application.NewTradeCardApplication(service.NewTradeCardService(
		tradeCardRepository, transactionRepository, ledgerLoader, clockProxy, calendar))
	tradeOrderApplication := application.NewTradeOrderApplication(service.NewTradeOrderService(
		tradeOrderRepository, orderFillRepository, transactionRepository, ledgerLoader, clockProxy, calendar,
		domains.NewFeeScheduleDomain(applicationConfig.FeeSchedule)))
	portfolioApplication := application.NewPortfolioApplication(service.NewPortfolioService(ledgerLoader))

	errorPresenter := controller.NewErrorPresenter()
	callerContext := controller.NewCallerContext()

	return router.NewRouter(
		middlewares.NewCorsMiddleware(applicationConfig.CorsAllowedOrigins),
		middlewares.NewAuthenticationMiddleware(authenticationApplication, errorPresenter, callerContext),
		controller.NewAuthorizationController(authorizationApplication, errorPresenter, controller.NewRefreshTokenCookie(
			controller.RefreshCookieSettings{
				Domain: applicationConfig.RefreshCookie.Domain, IsSecure: applicationConfig.RefreshCookie.IsSecure,
			}, applicationConfig.CorsAllowedOrigins)),
		controller.NewAuthenticationController(authenticationApplication),
		controller.NewApiKeyController(apiKeyApplication, errorPresenter),
		controller.NewTradeCardController(tradeCardApplication, errorPresenter, callerContext),
		controller.NewTradeOrderController(tradeOrderApplication, errorPresenter, callerContext),
		controller.NewPortfolioController(portfolioApplication, errorPresenter),
		controller.NewHealthController(),
	)
}
