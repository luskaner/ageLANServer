package internal

import (
	"crypto/sha512"

	"github.com/r3labs/diff/v3"
)

import (
	"net/http"
)

type Response struct {
	Headers    http.Header `diff:"headers"`
	StatusCode int         `diff:"status_code"`
	Body       []byte      `diff:"body"`
	BodyHash   [64]byte    `diff:"body_hash"`
}

func NewResponseWithBody(headers http.Header, statusCode int, body []byte) *Response {
	response := NewResponseWithoutBody(headers, statusCode)
	if len(body) > 0 {
		response.Body = body
		response.BodyHash = sha512.Sum512(body)
	}
	return response
}

func NewResponseWithoutBody(headers http.Header, statusCode int) *Response {
	return &Response{
		Headers:    headers,
		StatusCode: statusCode,
	}
}

func (r *Response) Matches(another *Response) (diff.Changelog, error) {
	return diff.Diff(r, another)
}
