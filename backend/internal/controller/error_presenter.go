package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/CodeMachine0121/trading-board/backend/internal/controller/models"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/gin-gonic/gin"
)

// ErrorPresenter turns an error into the contract's error body and status; anything unrecognised is a 500
// whose details stay in the log, not in the response.
type ErrorPresenter struct{}

func NewErrorPresenter() *ErrorPresenter {
	return &ErrorPresenter{}
}

func (errorPresenter *ErrorPresenter) Respond(ginContext *gin.Context, err error) {
	var businessError domains.BusinessError
	if !errors.As(err, &businessError) {
		log.Printf("%s %s failed: %v", ginContext.Request.Method, ginContext.FullPath(), err)
		ginContext.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorBody{Code: "internalError", Message: "伺服器發生錯誤，請稍後再試"}})
		return
	}

	status := http.StatusBadRequest
	switch {
	case errors.Is(err, domains.ErrAuthenticationRequired):
		status = http.StatusUnauthorized
	case errors.Is(err, domains.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domains.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, domains.ErrForbidden):
		status = http.StatusForbidden
	}

	ginContext.AbortWithStatusJSON(status, models.ErrorResponse{
		Error: models.ErrorBody{Code: businessError.Code, Message: businessError.Message}})
}

// RespondUnreadableBody answers a body that is not even the right JSON shape.
func (errorPresenter *ErrorPresenter) RespondUnreadableBody(ginContext *gin.Context, bindError error) {
	errorPresenter.Respond(ginContext, domains.ValidationFailed("invalidRequest", "無法讀取送出的內容："+bindError.Error()))
}

// RespondOAuth answers the sign-in endpoints in OAuth's own shape: { "error", "error_description" }.
func (errorPresenter *ErrorPresenter) RespondOAuth(ginContext *gin.Context, err error) {
	var oauthError domains.OAuthError
	if !errors.As(err, &oauthError) {
		log.Printf("%s %s failed: %v", ginContext.Request.Method, ginContext.FullPath(), err)
		ginContext.AbortWithStatusJSON(http.StatusInternalServerError, models.OAuthErrorResponse{
			Error: "server_error", ErrorDescription: "伺服器發生錯誤，請稍後再試"})
		return
	}

	status := http.StatusBadRequest
	switch {
	case errors.Is(err, domains.ErrAuthenticationRequired):
		status = http.StatusUnauthorized
	case errors.Is(err, domains.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, domains.ErrTooManyAttempts):
		status = http.StatusTooManyRequests
	}

	ginContext.AbortWithStatusJSON(status, models.OAuthErrorResponse{
		Error: oauthError.Code, ErrorDescription: oauthError.Description})
}
