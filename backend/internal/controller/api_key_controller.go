package controller

import (
	"net/http"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/models"
	"github.com/gin-gonic/gin"
)

// ApiKeyController lets the signed-in owner manage the agent-team members' API keys.
type ApiKeyController struct {
	apiKeyApplication *application.ApiKeyApplication
	errorPresenter    *ErrorPresenter
}

func NewApiKeyController(apiKeyApplication *application.ApiKeyApplication, errorPresenter *ErrorPresenter) *ApiKeyController {
	return &ApiKeyController{apiKeyApplication: apiKeyApplication, errorPresenter: errorPresenter}
}

func (apiKeyController *ApiKeyController) ListApiKeys(ginContext *gin.Context) {
	apiKeyDtos, listError := apiKeyController.apiKeyApplication.List(ginContext.Request.Context())
	if listError != nil {
		apiKeyController.errorPresenter.Respond(ginContext, listError)
		return
	}

	ginContext.JSON(http.StatusOK, apiKeyDtos)
}

// CreateApiKey answers 201 with the full key, which is never shown again.
func (apiKeyController *ApiKeyController) CreateApiKey(ginContext *gin.Context) {
	var apiKeyRequest models.ApiKeyRequest
	if bindError := ginContext.ShouldBindJSON(&apiKeyRequest); bindError != nil {
		apiKeyController.errorPresenter.RespondUnreadableBody(ginContext, bindError)
		return
	}

	createdDto, createError := apiKeyController.apiKeyApplication.Create(
		ginContext.Request.Context(), apiKeyRequest.ToCreateDto())
	if createError != nil {
		apiKeyController.errorPresenter.Respond(ginContext, createError)
		return
	}

	ginContext.Header("Cache-Control", "no-store")
	ginContext.JSON(http.StatusCreated, createdDto)
}

func (apiKeyController *ApiKeyController) RevokeApiKey(ginContext *gin.Context) {
	apiKeyDto, revokeError := apiKeyController.apiKeyApplication.Revoke(
		ginContext.Request.Context(), ginContext.Param("apiKeyId"))
	if revokeError != nil {
		apiKeyController.errorPresenter.Respond(ginContext, revokeError)
		return
	}

	ginContext.JSON(http.StatusOK, apiKeyDto)
}
