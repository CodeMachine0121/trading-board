package controller

import (
	"net/http"
	"slices"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
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

// RefreshTokenCookie owns every rule about the refresh-token cookie: it is HttpOnly, SameSite=Strict and scoped to
// the OAuth endpoints, and it is only read for a request from one of the web app's origins (CSRF protection on top
// of SameSite). Browsers send Origin on every POST, so a request without one is refused too.
type RefreshTokenCookie struct {
	settings       RefreshCookieSettings
	trustedOrigins []string
}

func NewRefreshTokenCookie(settings RefreshCookieSettings, trustedOrigins []string) *RefreshTokenCookie {
	return &RefreshTokenCookie{settings: settings, trustedOrigins: trustedOrigins}
}

// Read answers the refresh token the browser sent (empty when none), or a refusal for a stranger's request.
func (refreshCookie *RefreshTokenCookie) Read(ginContext *gin.Context) (string, error) {
	origin := ginContext.GetHeader("Origin")
	if origin == "" || !slices.Contains(refreshCookie.trustedOrigins, origin) {
		return "", domains.ForbiddenOrigin("請求來源不被允許")
	}
	refreshToken, _ := ginContext.Cookie(RefreshTokenCookieName)

	return refreshToken, nil
}

func (refreshCookie *RefreshTokenCookie) Write(ginContext *gin.Context, refreshToken string, expiresAt time.Time) {
	http.SetCookie(ginContext.Writer, refreshCookie.cookie(refreshToken, expiresAt, 0))
}

func (refreshCookie *RefreshTokenCookie) Clear(ginContext *gin.Context) {
	http.SetCookie(ginContext.Writer, refreshCookie.cookie("", time.Time{}, -1))
}

// cookie is the one place the cookie's attributes are spelled out, so writing and clearing can never disagree
// (a browser only clears a cookie whose name, path and domain match).
func (refreshCookie *RefreshTokenCookie) cookie(value string, expiresAt time.Time, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    value,
		Path:     refreshTokenCookiePath,
		Domain:   refreshCookie.settings.Domain,
		Expires:  expiresAt,
		MaxAge:   maxAge,
		Secure:   refreshCookie.settings.IsSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
