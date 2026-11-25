package controller

import (
	"net/http"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/gin-gonic/gin"
)

type PortfolioController struct {
	portfolioApplication *application.PortfolioApplication
	errorPresenter       *ErrorPresenter
}

func NewPortfolioController(
	portfolioApplication *application.PortfolioApplication, errorPresenter *ErrorPresenter,
) *PortfolioController {
	return &PortfolioController{portfolioApplication: portfolioApplication, errorPresenter: errorPresenter}
}

func (portfolioController *PortfolioController) ListPositions(ginContext *gin.Context) {
	positionDtos, listError := portfolioController.portfolioApplication.ListPositions(ginContext.Request.Context(),
		dto.PositionQueryDto{Status: ginContext.Query("status"), Month: ginContext.Query("month")})
	if listError != nil {
		portfolioController.errorPresenter.Respond(ginContext, listError)
		return
	}

	ginContext.JSON(http.StatusOK, positionDtos)
}

func (portfolioController *PortfolioController) GetMonthSummary(ginContext *gin.Context) {
	summaryDto, summaryError := portfolioController.portfolioApplication.GetMonthSummary(
		ginContext.Request.Context(), ginContext.Param("month"))
	if summaryError != nil {
		portfolioController.errorPresenter.Respond(ginContext, summaryError)
		return
	}

	ginContext.JSON(http.StatusOK, summaryDto)
}
