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
	clockProxy := clock.NewSystemClockProxy()
	accessTokenProxy := security.NewJwtAccessTokenProxy(applicationConfig.AccessTokenSigningKey)
	calendar := domains.NewBusinessCalendarDomain(applicationConfig.BusinessTimeZone)

	ledgerLoader := service.NewTradeLedgerLoader(
		tradeOrderRepository, tradeCardRepository, clockProxy, calendar, applicationConfig.BudgetPolicy)
	authenticationApplication := application.NewAuthenticationApplication(service.NewAuthenticationService(
		domains.NewAccessCredentialDomain(applicationConfig.AccessCredential),
		applicationConfig.AccessTokenLifetime, accessTokenProxy, clockProxy, calendar))
	tradeCardApplication := application.NewTradeCardApplication(service.NewTradeCardService(
		tradeCardRepository, transactionRepository, ledgerLoader, clockProxy, calendar))
	tradeOrderApplication := application.NewTradeOrderApplication(service.NewTradeOrderService(
		tradeOrderRepository, orderFillRepository, transactionRepository, ledgerLoader, clockProxy, calendar,
		domains.NewFeeScheduleDomain(applicationConfig.FeeSchedule)))
	portfolioApplication := application.NewPortfolioApplication(service.NewPortfolioService(ledgerLoader))

	errorPresenter := controller.NewErrorPresenter()

	return router.NewRouter(
		middlewares.NewCorsMiddleware(applicationConfig.CorsAllowedOrigins),
		middlewares.NewAuthenticationMiddleware(authenticationApplication, errorPresenter),
		controller.NewAuthenticationController(authenticationApplication, errorPresenter),
		controller.NewTradeCardController(tradeCardApplication, errorPresenter),
		controller.NewTradeOrderController(tradeOrderApplication, errorPresenter),
		controller.NewPortfolioController(portfolioApplication, errorPresenter),
		controller.NewHealthController(),
	)
}
