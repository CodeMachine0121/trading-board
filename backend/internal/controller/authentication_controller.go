package controller

import (
	"net/http"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/CodeMachine0121/trading-board/backend/internal/controller/models"
	"github.com/gin-gonic/gin"
)

type AuthenticationController struct {
	authenticationApplication *application.AuthenticationApplication
	errorPresenter            *ErrorPresenter
}

func NewAuthenticationController(
	authenticationApplication *application.AuthenticationApplication, errorPresenter *ErrorPresenter,
) *AuthenticationController {
	return &AuthenticationController{authenticationApplication: authenticationApplication, errorPresenter: errorPresenter}
}

func (authenticationController *AuthenticationController) SignIn(ginContext *gin.Context) {
	var signInRequest models.SignInRequest
	if bindError := ginContext.ShouldBindJSON(&signInRequest); bindError != nil {
		authenticationController.errorPresenter.RespondUnreadableBody(ginContext, bindError)
		return
	}

	accessTokenDto, signInError := authenticationController.authenticationApplication.SignIn(
		ginContext.Request.Context(), signInRequest.ToSignInDto())
	if signInError != nil {
		authenticationController.errorPresenter.Respond(ginContext, signInError)
		return
	}

	ginContext.JSON(http.StatusOK, accessTokenDto)
}
