package middlewares

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// CorsMiddleware lets the configured frontends read responses from another origin.
type CorsMiddleware struct {
	allowedOrigins []string
}

func NewCorsMiddleware(allowedOrigins []string) *CorsMiddleware {
	return &CorsMiddleware{allowedOrigins: allowedOrigins}
}

func (corsMiddleware *CorsMiddleware) Handle(ginContext *gin.Context) {
	origin := ginContext.GetHeader("Origin")
	if origin != "" && slices.Contains(corsMiddleware.allowedOrigins, origin) {
		headers := ginContext.Writer.Header()
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		headers.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Api-Key")
		headers.Set("Access-Control-Max-Age", "600")
		headers.Add("Vary", "Origin")
	}

	if ginContext.Request.Method == http.MethodOptions {
		ginContext.AbortWithStatus(http.StatusNoContent)
		return
	}
	ginContext.Next()
}
