package main

import (
	"ashno-onepay/internal/config"
	"ashno-onepay/internal/controller"
	"ashno-onepay/internal/jwt"
	"ashno-onepay/internal/log"
	middleware "ashno-onepay/internal/middleware"
	"ashno-onepay/internal/repository"
	"ashno-onepay/internal/server"
	"ashno-onepay/internal/service"
	"runtime"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	config.InitDatabase()

	cfg := config.GetConfig()

	logger := log.New(log.Config{
		Level: cfg.Log.GetLogLever(),
	})

	// Setting the default time zone for the application to UTC
	time.Local, _ = time.LoadLocation(cfg.Database.TimeZone)

	http := server.NewHTTPServer(logger)

	//repo
	userRepo := repository.GetUserRepositoryInstance(config.GetDB())
	//service
	userSvc := service.GetUserServiceInstance(userRepo)
	//controller
	jwtIssuer := jwt.NewIssuer(cfg.Server.JwtKey)
	userCtrl := controller.NewUserController(jwtIssuer, userSvc)

	sessionMiddleware := middleware.NewSessionMiddleware(jwt.NewValidator(cfg.Server.JwtKey))

	sv := server.NewServer(
		logger, &cfg, http,
		userCtrl,
		sessionMiddleware)
	sv.Run()

}
