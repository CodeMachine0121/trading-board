package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthController answers the platform's probes; it needs no credentials and touches nothing.
type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (healthController *HealthController) Check(ginContext *gin.Context) {
	ginContext.JSON(http.StatusOK, gin.H{"status": "ok"})
}
