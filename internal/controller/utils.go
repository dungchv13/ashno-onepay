package controller

import (
	"ashno-onepay/internal/errors"
	"ashno-onepay/internal/trace"
	"github.com/gin-gonic/gin"
)

func handleError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}
	err = errors.AppendTraceID(err, trace.GetTraceID(ctx))
	ctx.Error(err)

	if appError, ok := err.(errors.AppError); ok {
		ctx.JSON(appError.StatusCode, err)
	} else {
		err := errors.ErrInternal.Wrap(err).Reform(err.Error())
		err = err.SetTraceID(trace.GetTraceID(ctx))
		ctx.JSON(errors.ErrInternal.StatusCode, err)
	}
}
