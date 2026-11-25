package application

import (
	"context"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/service"
)

type AuthorizationApplication struct {
	authorizationService *service.AuthorizationService
}

func NewAuthorizationApplication(authorizationService *service.AuthorizationService) *AuthorizationApplication {
	return &AuthorizationApplication{authorizationService: authorizationService}
}

func (authorizationApplication *AuthorizationApplication) Authorize(
	executionContext context.Context, requestDto dto.AuthorizationRequestDto,
) (dto.AuthorizationRedirectDto, error) {
	return authorizationApplication.authorizationService.Authorize(executionContext, requestDto)
}

func (authorizationApplication *AuthorizationApplication) ExchangeCode(
	executionContext context.Context, exchangeDto dto.CodeExchangeDto,
) (dto.TokenSetDto, error) {
	return authorizationApplication.authorizationService.ExchangeCode(executionContext, exchangeDto)
}

func (authorizationApplication *AuthorizationApplication) Refresh(
	executionContext context.Context, refreshToken string,
) (dto.TokenSetDto, error) {
	return authorizationApplication.authorizationService.Refresh(executionContext, refreshToken)
}

func (authorizationApplication *AuthorizationApplication) Revoke(
	executionContext context.Context, refreshToken string,
) error {
	return authorizationApplication.authorizationService.Revoke(executionContext, refreshToken)
}
