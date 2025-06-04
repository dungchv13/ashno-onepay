package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "DELETE", "OPTIONS", "PUT", "PATCH"}
	corsConfig.AllowHeaders = []string{
		"RecaptchaToken",
		"AccessToken",
		"Authorization",
		"Content-Type",
		"Upgrade",
		"Origin",
		"Connection",
		"Accept-Encoding",
		"Accept-Language",
		"Host",
		"Access-Control-Request-Method",
		"Access-Control-Request-Headers",
		"official-account-id",
		"x-xss-protection",
	}
	return cors.New(corsConfig)
}
