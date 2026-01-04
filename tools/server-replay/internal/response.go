package internal

import (
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

func (r *Response) Matches(another *Response) (diff.Changelog, error) {
	return diff.Diff(r, another)
}
