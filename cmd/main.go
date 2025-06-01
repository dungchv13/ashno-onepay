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
	registrationRepo := repository.GetRegistrationRepositoryInstance(config.GetDB())
	registrationOptionsRepo := repository.GetRegistrationOptionRepositoryInstance(config.GetDB())
	//service
	registrationSvc := service.GetRegistrationServiceInstance(registrationRepo, registrationOptionsRepo, &cfg)
	//controller
	registrationCtrl := controller.NewRegistrationController(registrationSvc)

	sessionMiddleware := middleware.NewSessionMiddleware(jwt.NewValidator(cfg.Server.JwtKey))

	sv := server.NewServer(
		logger, &cfg, http,
		registrationCtrl,
		sessionMiddleware)
	sv.Run()

}
