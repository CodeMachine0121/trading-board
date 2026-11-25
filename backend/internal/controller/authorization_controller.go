package controller

import (
	"net/http"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/models"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/gin-gonic/gin"
)

// AuthorizationController speaks OAuth 2.0 on the wire: the login form, the token endpoint and revocation.
// The refresh token never appears in a body; RefreshTokenCookie carries it.
type AuthorizationController struct {
	authorizationApplication *application.AuthorizationApplication
	errorPresenter           *ErrorPresenter
	refreshCookie            *RefreshTokenCookie
}

func NewAuthorizationController(
	authorizationApplication *application.AuthorizationApplication,
	errorPresenter *ErrorPresenter,
	refreshCookie *RefreshTokenCookie,
) *AuthorizationController {
	return &AuthorizationController{
		authorizationApplication: authorizationApplication,
		errorPresenter:           errorPresenter,
		refreshCookie:            refreshCookie,
	}
}

func (authorizationController *AuthorizationController) Authorize(ginContext *gin.Context) {
	ginContext.Header("Cache-Control", "no-store")
	var authorizeRequest models.AuthorizeRequest
	if bindError := ginContext.ShouldBindJSON(&authorizeRequest); bindError != nil {
		authorizationController.errorPresenter.RespondOAuth(ginContext, domains.InvalidRequest("無法讀取送出的內容"))
		return
	}

	redirectDto, authorizeError := authorizationController.authorizationApplication.Authorize(
		ginContext.Request.Context(), authorizeRequest.ToAuthorizationRequestDto())
	if authorizeError != nil {
		authorizationController.errorPresenter.RespondOAuth(ginContext, authorizeError)
		return
	}

	ginContext.JSON(http.StatusOK, redirectDto)
}

func (authorizationController *AuthorizationController) Token(ginContext *gin.Context) {
	ginContext.Header("Cache-Control", "no-store")
	ginContext.Header("Pragma", "no-cache")
	var tokenRequest models.TokenRequest
	if bindError := ginContext.ShouldBind(&tokenRequest); bindError != nil {
		authorizationController.errorPresenter.RespondOAuth(ginContext, domains.InvalidRequest("無法讀取送出的內容"))
		return
	}

	var tokenSet dto.TokenSetDto
	var grantError error
	switch tokenRequest.GrantType {
	case "authorization_code":
		tokenSet, grantError = authorizationController.authorizationApplication.ExchangeCode(
			ginContext.Request.Context(), tokenRequest.ToCodeExchangeDto())
	case "refresh_token":
		refreshToken, readError := authorizationController.refreshCookie.Read(ginContext)
		if readError != nil {
			authorizationController.errorPresenter.RespondOAuth(ginContext, readError)
			return
		}
		tokenSet, grantError = authorizationController.authorizationApplication.Refresh(
			ginContext.Request.Context(), refreshToken)
		if grantError != nil {
			authorizationController.refreshCookie.Clear(ginContext)
		}
	default:
		grantError = domains.UnsupportedGrantType("只支援 authorization_code 與 refresh_token")
	}
	if grantError != nil {
		authorizationController.errorPresenter.RespondOAuth(ginContext, grantError)
		return
	}

	authorizationController.refreshCookie.Write(ginContext, tokenSet.RefreshToken, tokenSet.RefreshTokenExpiresAt)
	ginContext.JSON(http.StatusOK, tokenSet)
}

// Revoke signs out (RFC 7009): it always answers 200 for an accepted request and forgets the cookie.
func (authorizationController *AuthorizationController) Revoke(ginContext *gin.Context) {
	ginContext.Header("Cache-Control", "no-store")
	refreshToken, readError := authorizationController.refreshCookie.Read(ginContext)
	if readError != nil {
		authorizationController.errorPresenter.RespondOAuth(ginContext, readError)
		return
	}

	if revokeError := authorizationController.authorizationApplication.Revoke(
		ginContext.Request.Context(), refreshToken); revokeError != nil {
		authorizationController.errorPresenter.RespondOAuth(ginContext, revokeError)
		return
	}

	authorizationController.refreshCookie.Clear(ginContext)
	ginContext.Status(http.StatusOK)
}
