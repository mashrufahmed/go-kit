package httpx

import (
	"net/http"
	"testing"
)

func TestHandler(t *testing.T) {
	var handler Handler = func(w Res, r *Req) error {
		return nil
	}

	if handler == nil {
		t.Fatal("handler should not be nil")
	}

	_ = http.MethodGet
}
