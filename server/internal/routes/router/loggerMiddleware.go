package router

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/textproto"
	"time"

	"github.com/luskaner/ageLANServer/server/internal/logger"
)

var keepRequestHeaders = []string{
	"host",
	"content-type",
	"content-length",
	"transfer-encoding",
	"connection",
	"upgrade",
	"proxy-connection",
	"accept-encoding",
	"sec-websocket-key",
	"sec-websocket-version",
	"authorization",
	"x-entitytoken",
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Body:           new(bytes.Buffer),
	}
}

func (w *ResponseWriterWrapper) WriteHeader(code int) {
	if w.StatusCode == http.StatusOK {
		w.StatusCode = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *ResponseWriterWrapper) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *ResponseWriterWrapper) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

func encodeBase64(body []byte) string {
	return base64.StdEncoding.EncodeToString(body)
}

func filterHeaders(original http.Header) http.Header {
	filtered := make(http.Header)
	for _, key := range keepRequestHeaders {
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		values := original[canonicalKey]
		if len(values) > 0 {
			filtered[canonicalKey] = values
		}
	}
	return filtered
}

func NewLoggingMiddleware(next http.Handler, t time.Time) http.Handler {
	logger.StartTime = t
	logger.SlogEnabled = true
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestStart := time.Now()
		requestBody, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		wWrapper := NewResponseWriterWrapper(w)
		next.ServeHTTP(wWrapper, r)
		requestLatency := time.Since(requestStart)
		responseBody := wWrapper.Body.Bytes()
		requestGroup := slog.Group("in",
			slog.String("host", r.Host),
			slog.String("method", r.Method),
			slog.String("path", r.RequestURI),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Any("headers", filterHeaders(r.Header)),
			slog.String("body", encodeBase64(requestBody)),
		)
		bodyHash := ""
		if len(responseBody) > 0 {
			hash := sha512.Sum512(responseBody)
			bodyHash = encodeBase64(hash[:])
		}
		responseBodyStr := ""
		if len(responseBody) <= 4_096 {
			responseBodyStr = encodeBase64(responseBody)
		}
		responseGroup := slog.Group("out",
			slog.Int("status_code", wWrapper.StatusCode),
			slog.Int64("latency", requestLatency.Milliseconds()),
			slog.Any("headers", wWrapper.Header()),
			slog.String("body", responseBodyStr),
			slog.String("body_hash", bodyHash),
		)
		logger.LogMessage("request", requestGroup, responseGroup)
	})
}
