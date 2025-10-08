package test

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func Test(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{})
}
