package test

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
)

func Test(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{})
}
