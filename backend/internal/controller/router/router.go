package router

import (
	"github.com/CodeMachine0121/trading-board/backend/internal/controller"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/middlewares"
	"github.com/gin-gonic/gin"
)

// Router is the route table of the shared contract; it is the only place that says who may call what.
type Router struct {
	corsMiddleware           *middlewares.CorsMiddleware
	authenticationMiddleware *middlewares.AuthenticationMiddleware
	authorizationController  *controller.AuthorizationController
	authenticationController *controller.AuthenticationController
	apiKeyController         *controller.ApiKeyController
	tradeCardController      *controller.TradeCardController
	tradeOrderController     *controller.TradeOrderController
	portfolioController      *controller.PortfolioController
	healthController         *controller.HealthController
}

func NewRouter(
	corsMiddleware *middlewares.CorsMiddleware,
	authenticationMiddleware *middlewares.AuthenticationMiddleware,
	authorizationController *controller.AuthorizationController,
	authenticationController *controller.AuthenticationController,
	apiKeyController *controller.ApiKeyController,
	tradeCardController *controller.TradeCardController,
	tradeOrderController *controller.TradeOrderController,
	portfolioController *controller.PortfolioController,
	healthController *controller.HealthController,
) *Router {
	return &Router{
		corsMiddleware:           corsMiddleware,
		authenticationMiddleware: authenticationMiddleware,
		authorizationController:  authorizationController,
		authenticationController: authenticationController,
		apiKeyController:         apiKeyController,
		tradeCardController:      tradeCardController,
		tradeOrderController:     tradeOrderController,
		portfolioController:      portfolioController,
		healthController:         healthController,
	}
}

func (router *Router) Register(engine *gin.Engine) {
	engine.Use(router.corsMiddleware.Handle)
	engine.GET("/health", router.healthController.Check)

	api := engine.Group("/api/v1")
	api.GET("/health", router.healthController.Check)
	api.POST("/oauth/authorize", router.authorizationController.Authorize)
	api.POST("/oauth/token", router.authorizationController.Token)
	api.POST("/oauth/revoke", router.authorizationController.Revoke)

	ownerOnly := api.Group("", router.authenticationMiddleware.RequireOwner)
	ownerOnly.GET("/me", router.authenticationController.GetCurrentOwner)
	ownerOnly.GET("/api-keys", router.apiKeyController.ListApiKeys)
	ownerOnly.POST("/api-keys", router.apiKeyController.CreateApiKey)
	ownerOnly.POST("/api-keys/:apiKeyId/revoke", router.apiKeyController.RevokeApiKey)

	api.PUT("/cards/:tradeDate", router.authenticationMiddleware.RequireAgent, router.tradeCardController.PublishCard)

	readable := api.Group("", router.authenticationMiddleware.RequireOwnerOrAgent)
	readable.GET("/cards", router.tradeCardController.ListCards)
	readable.GET("/cards/latest", router.tradeCardController.GetLatestCard)
	readable.GET("/cards/:tradeDate", router.tradeCardController.GetCard)
	readable.POST("/orders/:orderId/fills", router.tradeOrderController.ReportFill)
	readable.POST("/orders/:orderId/skip", router.tradeOrderController.SkipOrder)
	readable.GET("/positions", router.portfolioController.ListPositions)
	readable.GET("/months/:month/summary", router.portfolioController.GetMonthSummary)
}
