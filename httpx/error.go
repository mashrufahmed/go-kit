package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mashrufahmed/go-kit/validate"
)

type HTTPError struct {
	Status    int
	Code      string
	Message   string
	Details   any
	Err       error
	Timestamp time.Time
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	if e.Err != nil {
		return e.Err.Error()
	}

	return fmt.Sprintf(
		"HTTP error: %d",
		e.Status,
	)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func NewError(
	status int,
	code string,
	message string,
) *HTTPError {
	return &HTTPError{
		Status:    status,
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
	}
}

func BadRequest(message string) error {
	return NewError(
		http.StatusBadRequest,
		"BAD_REQUEST",
		message,
	)
}

func Unauthorized(message string) error {
	return NewError(
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		message,
	)
}

func Forbidden(message string) error {
	return NewError(
		http.StatusForbidden,
		"FORBIDDEN",
		message,
	)
}

func NotFound(message string) error {
	return NewError(
		http.StatusNotFound,
		"NOT_FOUND",
		message,
	)
}

func Conflict(message string) error {
	return NewError(
		http.StatusConflict,
		"CONFLICT",
		message,
	)
}

func InternalServerError(message string) error {
	return NewError(
		http.StatusInternalServerError,
		"INTERNAL_SERVER_ERROR",
		message,
	)
}

func writeError(
	w Res,
	err error,
) {
	var httpErr *HTTPError

	switch {
	case errors.Is(err, ErrEmptyBody):
		httpErr = NewError(
			http.StatusBadRequest,
			"EMPTY_BODY",
			"Request body is required",
		)

	case errors.Is(err, ErrInvalidJSON):
		httpErr = NewError(
			http.StatusBadRequest,
			"INVALID_JSON",
			"Request body contains invalid JSON",
		)
	case errors.Is(err, ErrUnsupportedContentType):
		httpErr = NewError(http.StatusUnsupportedMediaType, "UNSUPPORTED_CONTENT_TYPE", "Content-Type must be application/json")

	case validate.IsValidationError(err):
		httpErr = NewError(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
		)

		httpErr.Details = validate.Details(err)

	case errors.As(err, &httpErr):
		// Already an HTTPError.

	default:
		httpErr = NewError(
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"Internal server error",
		)
	}
	if httpErr.Status < 400 || httpErr.Status > 599 {
		httpErr.Status = http.StatusInternalServerError
	}
	if httpErr.Timestamp.IsZero() {
		httpErr.Timestamp = time.Now().UTC()
	}

	errorBody := map[string]any{
		"code":      httpErr.Code,
		"message":   httpErr.Message,
		"timestamp": httpErr.Timestamp,
	}

	if httpErr.Details != nil {
		errorBody["details"] = httpErr.Details
	}

	_ = JSON(
		w,
		httpErr.Status,
		map[string]any{
			"error": errorBody,
		},
	)
}
