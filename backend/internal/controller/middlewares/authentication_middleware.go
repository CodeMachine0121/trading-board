package middlewares

import (
	"strings"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/gin-gonic/gin"
)

const (
	bearerScheme = "bearer "
	apiKeyHeader = "X-Api-Key"
)

// AuthenticationMiddleware lets through the owner's pass or the agent key, as each route allows.
type AuthenticationMiddleware struct {
	authenticationApplication *application.AuthenticationApplication
	errorPresenter            *controller.ErrorPresenter
}

func NewAuthenticationMiddleware(
	authenticationApplication *application.AuthenticationApplication, errorPresenter *controller.ErrorPresenter,
) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		authenticationApplication: authenticationApplication,
		errorPresenter:            errorPresenter,
	}
}

// RequireOwnerOrAgent admits the signed-in owner and the agent team alike.
func (authenticationMiddleware *AuthenticationMiddleware) RequireOwnerOrAgent(ginContext *gin.Context) {
	if _, identifyError := authenticationMiddleware.identify(ginContext); identifyError != nil {
		authenticationMiddleware.errorPresenter.Respond(ginContext, identifyError)
		return
	}
	ginContext.Next()
}

// RequireAgent admits only the agent key; the owner's pass alone is not enough to publish.
func (authenticationMiddleware *AuthenticationMiddleware) RequireAgent(ginContext *gin.Context) {
	callerDto, identifyError := authenticationMiddleware.identify(ginContext)
	if identifyError == nil && callerDto.Role != string(vo.CallerRoleAgent) {
		identifyError = domains.AuthenticationRequired("只有顧問團隊的 API 金鑰可以發布交易卡")
	}
	if identifyError != nil {
		authenticationMiddleware.errorPresenter.Respond(ginContext, identifyError)
		return
	}
	ginContext.Next()
}

func (authenticationMiddleware *AuthenticationMiddleware) identify(ginContext *gin.Context) (dto.CallerDto, error) {
	accessToken := ""
	authorization := ginContext.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(authorization), bearerScheme) {
		accessToken = strings.TrimSpace(authorization[len(bearerScheme):])
	}

	return authenticationMiddleware.authenticationApplication.IdentifyCaller(ginContext.Request.Context(),
		dto.CallerCredentialDto{ApiKey: ginContext.GetHeader(apiKeyHeader), AccessToken: accessToken})
}
