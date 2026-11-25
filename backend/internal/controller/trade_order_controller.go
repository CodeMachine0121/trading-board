package controller

import (
	"net/http"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/models"
	"github.com/gin-gonic/gin"
)

type TradeOrderController struct {
	tradeOrderApplication *application.TradeOrderApplication
	errorPresenter        *ErrorPresenter
}

func NewTradeOrderController(
	tradeOrderApplication *application.TradeOrderApplication, errorPresenter *ErrorPresenter,
) *TradeOrderController {
	return &TradeOrderController{tradeOrderApplication: tradeOrderApplication, errorPresenter: errorPresenter}
}

func (orderController *TradeOrderController) ReportFill(ginContext *gin.Context) {
	var fillRequest models.OrderFillRequest
	if bindError := ginContext.ShouldBindJSON(&fillRequest); bindError != nil {
		orderController.errorPresenter.RespondUnreadableBody(ginContext, bindError)
		return
	}

	orderDto, reportError := orderController.tradeOrderApplication.ReportFill(
		ginContext.Request.Context(), ginContext.Param("orderId"), fillRequest.ToReportDto())
	if reportError != nil {
		orderController.errorPresenter.Respond(ginContext, reportError)
		return
	}

	ginContext.JSON(http.StatusOK, orderDto)
}

func (orderController *TradeOrderController) SkipOrder(ginContext *gin.Context) {
	orderDto, skipError := orderController.tradeOrderApplication.SkipOrder(
		ginContext.Request.Context(), ginContext.Param("orderId"))
	if skipError != nil {
		orderController.errorPresenter.Respond(ginContext, skipError)
		return
	}

	ginContext.JSON(http.StatusOK, orderDto)
}
