package controller

import (
	"net/http"

	"github.com/CodeMachine0121/trading-board/backend/internal/application"
	"github.com/gin-gonic/gin"
)

// AuthenticationController answers who is signed in.
type AuthenticationController struct {
	authenticationApplication *application.AuthenticationApplication
}

func NewAuthenticationController(authenticationApplication *application.AuthenticationApplication) *AuthenticationController {
	return &AuthenticationController{authenticationApplication: authenticationApplication}
}

func (authenticationController *AuthenticationController) GetCurrentOwner(ginContext *gin.Context) {
	ginContext.JSON(http.StatusOK, authenticationController.authenticationApplication.CurrentOwner())
}
