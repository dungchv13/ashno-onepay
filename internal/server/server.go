package server

import (
	"ashno-onepay/internal/config"
	"ashno-onepay/internal/controller"
	"ashno-onepay/internal/errors"
	"ashno-onepay/internal/log"
	"ashno-onepay/internal/middleware"
	"ashno-onepay/internal/trace"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

type Server struct {
	logger     log.Logger
	Config     *config.Config
	HTTPServer *gin.Engine
}

func NewServer(
	logger log.Logger,
	config *config.Config,
	httpServer *gin.Engine,
	registrationController *controller.RegistrationController,
	sessionMiddleware gin.HandlerFunc,
) *Server {
	httpServer.Use(func(ctx *gin.Context) {
		trace.AppendTraceID(ctx)
		start := time.Now()
		defer func() {
			if panicErr := recover(); panicErr != nil {
				body, _ := io.ReadAll(ctx.Request.Body)
				fields := log.Fields{
					"protocol":     "ashno",
					"elapsed_time": time.Since(start) / time.Millisecond,
					"request_body": string(body),
					"trace_id":     trace.GetTraceID(ctx),
				}
				fields = appendErrorFields(fields, fmt.Errorf("panic: %v", panicErr))
				if debug.Stack() != nil {
					fields = appendStacktraceFields(fields, debug.Stack())
				}
				logger.WithFields(fields).Info("panic")

				err := errors.ErrInternal.Reform("internal error : panic")
				err = err.AppendTraceID(trace.GetTraceID(ctx))
				ctx.Error(err)
				ctx.JSON(err.StatusCode, err)
				return
			}
		}()
		ctx.Header(trace.HeaderTraceKey, trace.GetTraceID(ctx))

		log.SetTraceIDInLogger(logger, ctx)
		ctx.Next()
	})

	AddSwagger(httpServer)

	//route
	{
		route := httpServer.Group("/")
		{
			route.POST("/register", registrationController.HandleRegister)
			route.GET("/register/:registerID/registration-info", registrationController.HandlerGetRegistrationInfo)
			route.GET("/onepay/ipn", registrationController.HandlerOnePayIPN)
			route.GET("/register/option", registrationController.HandlerGetOption)
		}
	}

	return &Server{
		logger:     logger,
		Config:     config,
		HTTPServer: httpServer,
	}
}

func (s *Server) Run() {
	sigint := make(chan os.Signal, 1)

	srv := &http.Server{
		Addr:              ":" + s.Config.Server.Port,
		Handler:           s.HTTPServer,
		ReadHeaderTimeout: 3 * time.Second,
	}
	go func() {
		s.logger.Infof("Server is running on port [port=%s]", s.Config.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("server is running error")
		}
	}()

	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("server forced to shutdown")
	}
	s.logger.Info("try to graceful shutdown")
	s.logger.Info("server exiting")

}

func NewHTTPServer(logger log.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(
		middleware.Cors(),
	)

	return engine
}

func corsMiddleware(c *gin.Context) {
	// Access-Control-Allow-Origin
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Next()
}

func appendErrorFields(
	fields log.Fields, err error,
) log.Fields {
	fields["error"] = err.Error()
	fields["cause_error"] = errors.Cause(err)
	return fields
}

func appendStacktraceFields(
	fields log.Fields, stackTrace []byte,
) log.Fields {
	fields["stack_trace"] = string(stackTrace)
	return fields
}
