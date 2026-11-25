package controller

import (
	"net/http"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/models"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/gin-gonic/gin"
)

type TradeCardController struct {
	tradeCardApplication *application.TradeCardApplication
	errorPresenter       *ErrorPresenter
}

func NewTradeCardController(
	tradeCardApplication *application.TradeCardApplication, errorPresenter *ErrorPresenter,
) *TradeCardController {
	return &TradeCardController{tradeCardApplication: tradeCardApplication, errorPresenter: errorPresenter}
}

func (cardController *TradeCardController) PublishCard(ginContext *gin.Context) {
	var cardRequest models.TradeCardRequest
	if bindError := ginContext.ShouldBindJSON(&cardRequest); bindError != nil {
		cardController.errorPresenter.RespondUnreadableBody(ginContext, bindError)
		return
	}

	tradeDate := ginContext.Param("tradeDate")
	if cardRequest.TradeDate != "" && cardRequest.TradeDate != tradeDate {
		cardController.errorPresenter.Respond(ginContext,
			domains.ValidationFailed("invalidRequest", "內容的 tradeDate 與網址上的交易日不一致"))
		return
	}

	cardDto, publishError := cardController.tradeCardApplication.PublishCard(
		ginContext.Request.Context(), cardRequest.ToPublishDto(tradeDate))
	if publishError != nil {
		cardController.errorPresenter.Respond(ginContext, publishError)
		return
	}

	ginContext.JSON(http.StatusOK, cardDto)
}

func (cardController *TradeCardController) ListCards(ginContext *gin.Context) {
	summaryDtos, listError := cardController.tradeCardApplication.ListCards(
		ginContext.Request.Context(), ginContext.Query("month"))
	if listError != nil {
		cardController.errorPresenter.Respond(ginContext, listError)
		return
	}

	ginContext.JSON(http.StatusOK, summaryDtos)
}

func (cardController *TradeCardController) GetLatestCard(ginContext *gin.Context) {
	cardDto, findError := cardController.tradeCardApplication.GetLatestCard(ginContext.Request.Context())
	if findError != nil {
		cardController.errorPresenter.Respond(ginContext, findError)
		return
	}

	ginContext.JSON(http.StatusOK, cardDto)
}

func (cardController *TradeCardController) GetCard(ginContext *gin.Context) {
	cardDto, findError := cardController.tradeCardApplication.GetCard(
		ginContext.Request.Context(), ginContext.Param("tradeDate"))
	if findError != nil {
		cardController.errorPresenter.Respond(ginContext, findError)
		return
	}

	ginContext.JSON(http.StatusOK, cardDto)
}
