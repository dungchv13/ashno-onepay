package trace

import (
	"ashno-onepay/internal/uuid"
	"github.com/gin-gonic/gin"
)

const (
	TraceKey       = "trace_id"
	HeaderTraceKey = "X-Trace-Id"
)

func AppendTraceID(ctx *gin.Context) {
	ctx.Set(TraceKey, uuid.NewNoDash())
}

func GetTraceID(ctx *gin.Context) string {
	traceID, ok := ctx.Get(TraceKey)
	if !ok {
		return ""
	}
	sTraceID, ok := traceID.(string)
	if !ok {
		return ""
	}
	return sTraceID
}
