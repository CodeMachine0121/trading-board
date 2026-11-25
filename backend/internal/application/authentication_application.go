package application

import (
	"context"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/service"
)

type AuthenticationApplication struct {
	authenticationService *service.AuthenticationService
}

func NewAuthenticationApplication(authenticationService *service.AuthenticationService) *AuthenticationApplication {
	return &AuthenticationApplication{authenticationService: authenticationService}
}

func (authenticationApplication *AuthenticationApplication) SignIn(
	executionContext context.Context, signInDto dto.SignInDto,
) (dto.AccessTokenDto, error) {
	return authenticationApplication.authenticationService.SignIn(executionContext, signInDto)
}

func (authenticationApplication *AuthenticationApplication) IdentifyCaller(
	executionContext context.Context, credentialDto dto.CallerCredentialDto,
) (dto.CallerDto, error) {
	return authenticationApplication.authenticationService.IdentifyCaller(executionContext, credentialDto)
}
