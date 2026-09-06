package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mashrufahmed/go-kit/validate"
)

var (
	ErrInvalidJSON            = errors.New("invalid JSON body")
	ErrEmptyBody              = errors.New("request body is empty")
	ErrUnsupportedContentType = errors.New("unsupported content type")
)

func Decode[T any](r *http.Request) (T, error) {
	var data T

	if r.Body == nil {
		return data, ErrEmptyBody
	}
	if contentType := r.Header.Get("Content-Type"); contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		return data, ErrUnsupportedContentType
	}

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&data); err != nil {
		if errors.Is(err, io.EOF) {
			return data, ErrEmptyBody
		}

		return data, fmt.Errorf(
			"%w: %v",
			ErrInvalidJSON,
			err,
		)
	}

	// Prevent multiple JSON values:
	// {} {}
	var extra any

	if err := decoder.Decode(&extra); err != io.EOF {
		return data, ErrInvalidJSON
	}

	return data, nil
}

func Param(
	r *http.Request,
	name string,
) string {
	return chi.URLParam(r, name)
}
func DecodeAndValidate[T any](r *http.Request) (T, error) {
	data, err := Decode[T](r)

	if err != nil {
		return data, err
	}

	if err := validate.Struct(data); err != nil {
		httpErr := BadRequest("Request validation failed").(*HTTPError)

		httpErr.Code = "VALIDATION_ERROR"
		httpErr.Details = validate.Details(data, err)

		return data, httpErr
	}

	return data, nil
}
