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

// AuthenticationMiddleware lets through the owner's access token or a member's API key, as each route allows,
// and leaves the identified caller on the request for attribution.
type AuthenticationMiddleware struct {
	authenticationApplication *application.AuthenticationApplication
	errorPresenter            *controller.ErrorPresenter
	callerContext             *controller.CallerContext
}

func NewAuthenticationMiddleware(
	authenticationApplication *application.AuthenticationApplication,
	errorPresenter *controller.ErrorPresenter,
	callerContext *controller.CallerContext,
) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		authenticationApplication: authenticationApplication,
		errorPresenter:            errorPresenter,
		callerContext:             callerContext,
	}
}

// RequireOwnerOrAgent admits the signed-in owner and every agent-team member alike.
func (authenticationMiddleware *AuthenticationMiddleware) RequireOwnerOrAgent(ginContext *gin.Context) {
	authenticationMiddleware.admit(ginContext, "", nil)
}

// RequireAgent admits only a member's API key; the owner's access token alone is not enough to publish.
func (authenticationMiddleware *AuthenticationMiddleware) RequireAgent(ginContext *gin.Context) {
	authenticationMiddleware.admit(ginContext, vo.CallerRoleAgent,
		domains.AuthenticationRequired("只有顧問團隊的 API 金鑰可以發布交易卡"))
}

// RequireOwner admits only the signed-in owner: an API key can never manage API keys.
func (authenticationMiddleware *AuthenticationMiddleware) RequireOwner(ginContext *gin.Context) {
	authenticationMiddleware.admit(ginContext, vo.CallerRoleOwner, domains.Forbidden("只有登入的擁有者可以使用"))
}

// admit identifies the caller and, when a role is required, answers wrongRole to anybody else.
func (authenticationMiddleware *AuthenticationMiddleware) admit(
	ginContext *gin.Context, requiredRole vo.CallerRoleVo, wrongRole error,
) {
	accessToken := ""
	authorization := ginContext.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(authorization), bearerScheme) {
		accessToken = strings.TrimSpace(authorization[len(bearerScheme):])
	}

	callerDto, identifyError := authenticationMiddleware.authenticationApplication.IdentifyCaller(
		ginContext.Request.Context(),
		dto.CallerCredentialDto{ApiKey: ginContext.GetHeader(apiKeyHeader), AccessToken: accessToken})
	if identifyError == nil && requiredRole != "" && callerDto.Role != string(requiredRole) {
		identifyError = wrongRole
	}
	if identifyError != nil {
		authenticationMiddleware.errorPresenter.Respond(ginContext, identifyError)
		return
	}

	authenticationMiddleware.callerContext.Remember(ginContext, callerDto)
	ginContext.Next()
}
