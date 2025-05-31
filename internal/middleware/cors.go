package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "DELETE", "PUT", "PATCH", "HEAD"}
	corsConfig.AllowHeaders = []string{
		"Authorization", "Content-Type", "Upgrade", "Origin", "Connection", "Accept-Encoding", "Accept-Language",
		"Host", "Access-Control-Request-Method", "Access-Control-Request-Headers", "session-key", "Api-Token",
		"GIMORGANIZATIONAPITOKEN",
	}
	return cors.New(corsConfig)
}
