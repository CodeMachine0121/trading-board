package controller

import (
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/gin-gonic/gin"
)

// callerKey is where the authentication middleware leaves the identified caller on the request.
const callerKey = "tradingBoardCaller"

// CallerContext carries who a request comes from from the authentication middleware to the handlers that record it.
type CallerContext struct{}

func NewCallerContext() *CallerContext {
	return &CallerContext{}
}

func (callerContext *CallerContext) Remember(ginContext *gin.Context, caller dto.CallerDto) {
	ginContext.Set(callerKey, caller)
}

// Caller answers an empty caller on a route the middleware does not guard; every guarded route has one.
func (callerContext *CallerContext) Caller(ginContext *gin.Context) dto.CallerDto {
	caller, _ := ginContext.Value(callerKey).(dto.CallerDto)

	return caller
}
