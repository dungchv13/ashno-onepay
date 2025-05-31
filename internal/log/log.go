package log

import (
	"ashno-onepay/internal/trace"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	ErrorKey = logrus.ErrorKey
)

const (
	TraceLoggerKey = "trace_logger"
)

type Logger = logrus.FieldLogger
type Fields = logrus.Fields

type Config struct {
	Level string
}

// Discord return no-op logger
func Discord() Logger {
	logger := logrus.New()
	logger.Out = io.Discard
	return logger
}

func New(c Config) Logger {
	logger := logrus.New()
	logger.Formatter = new(logrus.JSONFormatter)
	if c.Level != "" {
		lv, err := logrus.ParseLevel(c.Level)
		if err != nil {
			logger.WithError(err).Fatalf("log level parse failed [level=%s]", c.Level)
		}
		logger.Level = lv
	}

	return logger
}

func SetTraceIDInLogger(
	logger Logger,
	ctx *gin.Context,
) {
	logEntry := logger.WithFields(logrus.Fields{
		"trace_id": trace.GetTraceID(ctx),
	})
	ctx.Set(TraceLoggerKey, logEntry)
}

func GetLogger(ctx *gin.Context) Logger {
	ctxLogger, ok := ctx.Get(TraceLoggerKey)
	if !ok {
		return New(Config{
			Level: "info",
		})
	}
	logger, ok := ctxLogger.(Logger)
	if !ok {
		return New(Config{
			Level: "info",
		})
	}
	return logger
}
