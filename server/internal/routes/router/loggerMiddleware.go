package router

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/luskaner/ageLANServer/common/logger/serverCommunication"
	"github.com/luskaner/ageLANServer/common/logger/serverCommunication/request"
	"github.com/luskaner/ageLANServer/server/internal/logger"
)

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

func NewLoggingMiddleware(next http.Handler, t time.Time) http.Handler {
	logger.StartTime = t
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestStart := time.Now()
		requestBody, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		wWrapper := NewResponseWriterWrapper(w)
		next.ServeHTTP(wWrapper, r)
		requestLatency := time.Since(requestStart)
		responseBody := wWrapper.Body.Bytes()
		var hash [64]byte
		if r.Method != http.MethodHead && len(responseBody) > 0 {
			hash = sha512.Sum512(responseBody)
		}
		if len(responseBody) > 4_096 {
			responseBody = []byte{}
		}
		url := r.URL
		if r.URL.Scheme == "" {
			r.URL.Scheme = "https"
		}
		if r.URL.Host == "" {
			r.URL.Host = r.Host
		}
		req := request.NewWrite(request.Read{
			In: request.In{
				Base: request.Base{
					Body:    serverCommunication.Body{Body: requestBody},
					Headers: r.Header,
				},
				Uptime: serverCommunication.Uptime{Uptime: logger.Uptime(&requestStart)},
				Sender: serverCommunication.Sender{Sender: r.RemoteAddr},
				Url:    url,
				Method: r.Method,
			},
			Out: request.Out{
				Base: request.Base{
					Body:    serverCommunication.Body{Body: responseBody},
					Headers: wWrapper.Header(),
				},
				BodyHash:   serverCommunication.BodyHash{BodyHash: hash},
				StatusCode: wWrapper.StatusCode,
				Latency:    requestLatency,
			},
		})
		logger.CommBuffer.Log(req)
	})
}
