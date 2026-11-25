package controller

import (
	"net/http"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/models"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/gin-gonic/gin"
)

type TradeOrderController struct {
	tradeOrderApplication *application.TradeOrderApplication
	errorPresenter        *ErrorPresenter
	callerContext         *CallerContext
}

func NewTradeOrderController(
	tradeOrderApplication *application.TradeOrderApplication, errorPresenter *ErrorPresenter, callerContext *CallerContext,
) *TradeOrderController {
	return &TradeOrderController{
		tradeOrderApplication: tradeOrderApplication, errorPresenter: errorPresenter, callerContext: callerContext,
	}
}

func (orderController *TradeOrderController) ReportFill(ginContext *gin.Context) {
	var fillRequest models.OrderFillRequest
	if bindError := ginContext.ShouldBindJSON(&fillRequest); bindError != nil {
		orderController.errorPresenter.RespondUnreadableBody(ginContext, bindError)
		return
	}

	reportDto := fillRequest.ToReportDto()
	reportDto.ReportedBy = orderController.callerContext.Caller(ginContext).Name
	orderDto, reportError := orderController.tradeOrderApplication.ReportFill(
		ginContext.Request.Context(), ginContext.Param("orderId"), reportDto)
	if reportError != nil {
		orderController.errorPresenter.Respond(ginContext, reportError)
		return
	}

	ginContext.JSON(http.StatusOK, orderDto)
}

func (orderController *TradeOrderController) SkipOrder(ginContext *gin.Context) {
	orderDto, skipError := orderController.tradeOrderApplication.SkipOrder(ginContext.Request.Context(),
		ginContext.Param("orderId"), dto.OrderSkipDto{SkippedBy: orderController.callerContext.Caller(ginContext).Name})
	if skipError != nil {
		orderController.errorPresenter.Respond(ginContext, skipError)
		return
	}

	ginContext.JSON(http.StatusOK, orderDto)
}
