package server

import (
	"ashno-onepay/docs/swagger"
	"ashno-onepay/internal/config"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @securityDefinitions.apikey SessionKey
// @in header
// @name session-key
func AddSwagger(r *gin.Engine) {
	swgCfg := config.GetConfig().Swagger
	swagger.SwaggerInfo.Title = "ashno-onepay program"
	swagger.SwaggerInfo.Description = "ashno-onepay application"
	swagger.SwaggerInfo.Version = "1.0"
	swagger.SwaggerInfo.Host = swgCfg.Host
	swagger.SwaggerInfo.BasePath = swgCfg.BasePath
	swagger.SwaggerInfo.Schemes = swgCfg.GetSchemes()

	group := r.Group("/swagger", gin.BasicAuth(gin.Accounts{
		swgCfg.Username: swgCfg.Password,
	}))

	group.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
