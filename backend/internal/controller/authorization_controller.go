package controller

import (
	"net/http"
	"slices"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/models"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/gin-gonic/gin"
)

const (
	// RefreshTokenCookieName is the HttpOnly cookie holding the refresh token.
	RefreshTokenCookieName = "trading_board_refresh_token"
	// refreshTokenCookiePath keeps the cookie off every request except the token and revoke endpoints.
	refreshTokenCookiePath = "/api/v1/oauth"
)

// RefreshCookieSettings scopes the refresh-token cookie; an empty Domain means the API host only.
type RefreshCookieSettings struct {
	Domain   string
	IsSecure bool
}

// AuthorizationController speaks OAuth 2.0 on the wire: the login form, the token endpoint and revocation.
// The refresh token lives only in an HttpOnly, SameSite=Strict cookie, and any request that would use that
// cookie must come from one of the web app's origins (CSRF protection on top of SameSite).
type AuthorizationController struct {
	authorizationApplication *application.AuthorizationApplication
	errorPresenter           *ErrorPresenter
	cookieSettings           RefreshCookieSettings
	trustedOrigins           []string
}

func NewAuthorizationController(
	authorizationApplication *application.AuthorizationApplication,
	errorPresenter *ErrorPresenter,
	cookieSettings RefreshCookieSettings,
	trustedOrigins []string,
) *AuthorizationController {
	return &AuthorizationController{
		authorizationApplication: authorizationApplication,
		errorPresenter:           errorPresenter,
		cookieSettings:           cookieSettings,
		trustedOrigins:           trustedOrigins,
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
		if !authorizationController.isFromTrustedOrigin(ginContext) {
			authorizationController.errorPresenter.RespondOAuth(ginContext, domains.ForbiddenOrigin("請求來源不被允許"))
			return
		}
		refreshToken, _ := ginContext.Cookie(RefreshTokenCookieName)
		tokenSet, grantError = authorizationController.authorizationApplication.Refresh(
			ginContext.Request.Context(), refreshToken)
		if grantError != nil {
			authorizationController.clearRefreshCookie(ginContext)
		}
	default:
		grantError = domains.UnsupportedGrantType("只支援 authorization_code 與 refresh_token")
	}
	if grantError != nil {
		authorizationController.errorPresenter.RespondOAuth(ginContext, grantError)
		return
	}

	http.SetCookie(ginContext.Writer, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    tokenSet.RefreshToken,
		Path:     refreshTokenCookiePath,
		Domain:   authorizationController.cookieSettings.Domain,
		Expires:  tokenSet.RefreshTokenExpiresAt,
		Secure:   authorizationController.cookieSettings.IsSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	ginContext.JSON(http.StatusOK, tokenSet)
}

// Revoke signs out (RFC 7009): it always answers 200 for an accepted request and forgets the cookie.
func (authorizationController *AuthorizationController) Revoke(ginContext *gin.Context) {
	ginContext.Header("Cache-Control", "no-store")
	if !authorizationController.isFromTrustedOrigin(ginContext) {
		authorizationController.errorPresenter.RespondOAuth(ginContext, domains.ForbiddenOrigin("請求來源不被允許"))
		return
	}

	refreshToken, _ := ginContext.Cookie(RefreshTokenCookieName)
	if revokeError := authorizationController.authorizationApplication.Revoke(
		ginContext.Request.Context(), refreshToken); revokeError != nil {
		authorizationController.errorPresenter.RespondOAuth(ginContext, revokeError)
		return
	}

	authorizationController.clearRefreshCookie(ginContext)
	ginContext.Status(http.StatusOK)
}

// isFromTrustedOrigin requires the Origin header browsers send on every POST; a request without one is refused,
// since only the web app (a browser) ever holds the cookie.
func (authorizationController *AuthorizationController) isFromTrustedOrigin(ginContext *gin.Context) bool {
	origin := ginContext.GetHeader("Origin")

	return origin != "" && slices.Contains(authorizationController.trustedOrigins, origin)
}

func (authorizationController *AuthorizationController) clearRefreshCookie(ginContext *gin.Context) {
	http.SetCookie(ginContext.Writer, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     refreshTokenCookiePath,
		Domain:   authorizationController.cookieSettings.Domain,
		MaxAge:   -1,
		Secure:   authorizationController.cookieSettings.IsSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
