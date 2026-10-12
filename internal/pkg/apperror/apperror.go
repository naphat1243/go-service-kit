// Package apperror defines typed application errors and their mapping to
// HTTP responses, so handlers and services never hand-roll status codes.
package apperror

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Kind string

const (
	KindInvalid  Kind = "invalid_argument"
	KindNotFound Kind = "not_found"
	KindConflict Kind = "conflict"
	KindInternal Kind = "internal"
)

type Error struct {
	Kind    Kind
	Message string
	err     error
}

func (e *Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.err)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *Error) Unwrap() error { return e.err }

func Invalid(msg string) *Error  { return &Error{Kind: KindInvalid, Message: msg} }
func NotFound(msg string) *Error { return &Error{Kind: KindNotFound, Message: msg} }
func Conflict(msg string) *Error { return &Error{Kind: KindConflict, Message: msg} }
func Internal(msg string, err error) *Error {
	return &Error{Kind: KindInternal, Message: msg, err: err}
}

func statusOf(k Kind) int {
	switch k {
	case KindInvalid:
		return http.StatusBadRequest
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

type response struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// HTTPErrorHandler converts *Error (and echo's own errors) into a stable JSON
// envelope. Internal details are logged, never returned to clients.
func HTTPErrorHandler(log *slog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var appErr *Error
		if errors.As(err, &appErr) {
			if appErr.Kind == KindInternal {
				log.Error("internal error", "path", c.Path(), "err", err)
				_ = c.JSON(statusOf(appErr.Kind), response{Error: string(KindInternal), Message: "internal server error"})
				return
			}
			_ = c.JSON(statusOf(appErr.Kind), response{Error: string(appErr.Kind), Message: appErr.Message})
			return
		}

		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			_ = c.JSON(httpErr.Code, response{Error: http.StatusText(httpErr.Code), Message: fmt.Sprint(httpErr.Message)})
			return
		}

		log.Error("unhandled error", "path", c.Path(), "err", err)
		_ = c.JSON(http.StatusInternalServerError, response{Error: string(KindInternal), Message: "internal server error"})
	}
}
