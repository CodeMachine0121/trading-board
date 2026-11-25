package application

import (
	"context"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/service"
)

type ApiKeyApplication struct {
	apiKeyService *service.ApiKeyService
}

func NewApiKeyApplication(apiKeyService *service.ApiKeyService) *ApiKeyApplication {
	return &ApiKeyApplication{apiKeyService: apiKeyService}
}

func (apiKeyApplication *ApiKeyApplication) Create(
	executionContext context.Context, createDto dto.ApiKeyCreateDto,
) (dto.CreatedApiKeyDto, error) {
	return apiKeyApplication.apiKeyService.Create(executionContext, createDto)
}

func (apiKeyApplication *ApiKeyApplication) List(executionContext context.Context) ([]dto.ApiKeyDto, error) {
	return apiKeyApplication.apiKeyService.List(executionContext)
}

func (apiKeyApplication *ApiKeyApplication) Revoke(
	executionContext context.Context, apiKeyID string,
) (dto.ApiKeyDto, error) {
	return apiKeyApplication.apiKeyService.Revoke(executionContext, apiKeyID)
}
