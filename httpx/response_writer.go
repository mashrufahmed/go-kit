package httpx

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	written     int
	wroteHeader bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (w *responseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}

	w.status = status
	w.wroteHeader = true

	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(data)
	w.written += n

	return n, err
}

func (w *responseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		w.WriteHeader(http.StatusOK)
		f.Flush()
	}
}
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}
func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}
func (w *responseWriter) ReadFrom(r io.Reader) (int64, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		n, err := rf.ReadFrom(r)
		w.written += int(n)
		return n, err
	}
	n, err := io.Copy(w.ResponseWriter, r)
	w.written += int(n)
	return n, err
}
