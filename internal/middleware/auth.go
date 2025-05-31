package middleware

import (
	"ashno-onepay/internal/errors"
	"ashno-onepay/internal/jwt"
	"github.com/gin-gonic/gin"
)

const (
	CtxKeyCurrentUserClaims = "USER_CLAIMS"
	SessionKeyHeader        = "session-key"
)

func NewSessionMiddleware(validator jwt.Validator) func(c *gin.Context) {
	return func(c *gin.Context) {
		token := c.GetHeader(SessionKeyHeader)
		claims, err := validator.Validate(c.Request.Context(), token)
		if err != nil {
			handleAuthError(c, err)
			return
		}

		c.Set(CtxKeyCurrentUserClaims, claims)
		c.Next()
	}
}

func handleAuthError(ctx *gin.Context, err error) {
	handleError(ctx, err, errors.ErrUnauthorized)
}

func handleError(ctx *gin.Context, err error, defaultAppError errors.AppError) {
	if err == nil {
		return
	}
	ctx.Error(err)
	if appError, ok := err.(errors.AppError); ok {
		ctx.JSON(appError.StatusCode, err)
	} else {
		err := defaultAppError.Wrap(err)
		ctx.JSON(err.StatusCode, err)
	}
	ctx.Abort()
}
