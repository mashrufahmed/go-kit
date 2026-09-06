package httpx

import (
	"encoding/json"
	"net/http"
)

func JSON(
	w Res,
	status int,
	data any,
) error {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func OK(
	w Res,
	data any,
) error {
	return JSON(
		w,
		http.StatusOK,
		data,
	)
}

func Created(
	w Res,
	data any,
) error {
	return JSON(
		w,
		http.StatusCreated,
		data,
	)
}

func NoContent(
	w Res,
) error {
	w.WriteHeader(
		http.StatusNoContent,
	)

	return nil
}
