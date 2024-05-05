package test

import (
	"net/http"
	i "server/internal"
)

func Test(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{})
}
