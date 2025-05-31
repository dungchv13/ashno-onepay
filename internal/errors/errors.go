// nolint: stylecheck, lll
package errors

import (
	"net/http"

	"github.com/pkg/errors"
)

var (
	ErrBadRequest      = NewAppError(400, http.StatusBadRequest, "bad request")
	ErrInvalidArgument = NewAppError(400, http.StatusBadRequest, "invalid argument")

	ErrUnauthorized = NewAppError(400108, http.StatusBadRequest, "unauthorized request")
	ErrInvalidValue = NewAppError(400111, http.StatusBadRequest, "invalid value")
	ErrInternal     = NewAppError(500901, http.StatusInternalServerError, "internal error")

	ErrInvalidSession     = NewAppError(401, http.StatusUnauthorized, "invalid session")
	ErrInvalidIdentifier  = NewAppError(401, http.StatusUnauthorized, "invalid identifier")
	ErrInvalidPassword    = NewAppError(401, http.StatusUnauthorized, "invalid password")
	ErrForbidden          = NewAppError(403, http.StatusUnauthorized, "forbidden")
	ErrNotFound           = NewAppError(404, http.StatusNotFound, "not found")
	ErrOtherService       = NewAppError(7500001, http.StatusInternalServerError, "other service error")
	ErrDatabase           = NewAppError(7050004, http.StatusInternalServerError, "Server error")
	ErrHttpRequestTimeout = NewAppError(7500005, http.StatusRequestTimeout, "http request timeout")

	ErrInvalidRunTime = NewAppError(7500006, http.StatusBadRequest, "invalid run time")

	ErrParameterDecode = NewAppError(100000, http.StatusBadRequest, "parameter decode failure")
)

func New(msg string) error {
	return errors.New(msg)
}

func Wrap(err error, message string) error {
	var result error
	if appError, ok := err.(AppError); ok {
		appError.WrappedErr = errors.Wrap(appError.WrappedErr, message)
		result = appError
	} else {
		result = errors.Wrap(err, message)
	}

	return result
}

func AppendTraceID(err error, traceID string) error {
	var result error
	if appError, ok := err.(AppError); ok {
		result = appError.AppendTraceID(traceID)
	} else {
		result = errors.New(err.Error() + "( trace_id = " + traceID + " )")
	}

	return result
}

func Cause(err error) error {
	return errors.Cause(err)
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target error) bool {
	return errors.As(err, target)
}
