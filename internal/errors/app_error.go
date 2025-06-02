// nolint: stylecheck, lll
package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

type AppError struct {
	StatusCode    int    `json:"-"`
	ErrorIndicate bool   `json:"error"`
	Code          int    `json:"code"`
	Message       string `json:"message"`
	TraceID       string `json:"-"`

	WrappedErr error  `json:"-"`
	CallStack  string `json:"-"`
}

// stack Copy From errors plugin stack.go file
type stack []uintptr

func (e *AppError) callersStack() {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	stackString := make([]string, 0, 5)
	for i, pc := range st {
		if i < 5 {
			f := errors.Frame(pc)
			stack := fmt.Sprintf("%+v", f)
			split := strings.Split(stack, "\t")
			stackString = append(stackString, split[1])
		}
	}
	e.CallStack = strings.Join(stackString, "::")
}

// parsing to error type
func (g AppError) Error() string {
	data := map[string]interface{}{}

	if g.WrappedErr != nil {
		data["wrapped_error"] = g.WrappedErr.Error()
		data["cause"] = errors.Cause(g.WrappedErr).Error()
	}

	data["callStack"] = g.CallStack
	data["error"] = g.ErrorIndicate
	data["code"] = g.Code
	data["status_code"] = g.StatusCode
	data["message"] = g.Message
	data["trace_id"] = g.TraceID
	dataBytes, _ := json.Marshal(data)
	return string(dataBytes)
}

// ---------- func --------------------
func NewAppError(code int, status int, message string) AppError {
	return AppError{
		ErrorIndicate: true,
		Code:          code,
		StatusCode:    status,
		Message:       message,
	}
}

// reform the display message in error
func (e AppError) Reform(msg string, args ...interface{}) AppError {
	if len(msg) > 0 {
		e.callersStack()
		e.Message = fmt.Sprintf(msg, args...)
	}
	return e
}

func (e AppError) Wrap(err error) AppError {
	if err != nil {
		e.callersStack()
		e.WrappedErr = err
	}

	return e
}

func (e AppError) WrapString(err string) AppError {
	if len(err) != 0 {
		e.WrappedErr = errors.Wrap(e.WrappedErr, err)
	}

	return e
}

func (e AppError) AppendTraceID(traceID string) AppError {
	if len(traceID) > 0 {
		e.TraceID = traceID
		e.Message = fmt.Sprintf("%s", e.Message)
	}
	return e
}

func (e AppError) SetTraceID(traceID string) AppError {
	if len(traceID) > 0 {
		e.TraceID = traceID
	}
	return e
}
