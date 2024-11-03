package internal

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/schema"
	"net/http"
)

type A []any
type H map[string]any

var decoder = schema.NewDecoder()

func writeJSONHeader(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "application/json;charset=utf-8")
}

func JSON(w *http.ResponseWriter, data any) {
	writeJSONHeader(w)
	_ = json.NewEncoder(*w).Encode(data)
}

func RawJSON(w *http.ResponseWriter, data []byte) {
	writeJSONHeader(w)
	_, _ = (*w).Write(data)
}

func decode(dst interface{}, src map[string][]string) error {
	err := decoder.Decode(dst, src)
	if err == nil {
		return nil
	}

	var merr schema.MultiError
	if errors.As(err, &merr) {
		for k, err := range merr {
			var unknownKeyError schema.UnknownKeyError
			if errors.As(err, &unknownKeyError) {
				delete(merr, k)
			}
		}
		if len(merr) == 0 {
			return nil
		}
	}

	return err
}

func Bind[D any](r *http.Request, data *D) error {
	var err error
	if r.Method == http.MethodGet {
		err = decode(data, r.URL.Query())
	} else {
		err = r.ParseForm()
		if err != nil {
			return err
		}
		err = decode(data, r.PostForm)
	}
	return err
}
