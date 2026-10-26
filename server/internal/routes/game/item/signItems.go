package item

import (
	"net/http"

	i "github.com/luskaner/ageServer/server/internal"
	"github.com/luskaner/ageServer/server/internal/models"
)

type signItemsRequest struct {
	itemRequest
}

// SignItems answers the signature covering the requested items.
//
// The signed payload is the itemIDs value exactly as it arrived, so the client
// verifies against the very string it sent instead of against a re-serialisation
// of the same list. That only holds when this server signed the request, so with
// Sign off there is nothing to verify against and the reply is the same refusal
// the official server gives: result code 2 and no signature.
//
// A list that does not parse is refused the same way: there is no item set to
// speak of, so there is nothing to sign.
func SignItems(w http.ResponseWriter, r *http.Request) {
	if !i.SignConfigRuntime.Enabled {
		i.JSON(&w, i.A{2, ""})
		return
	}
	var req signItemsRequest
	if err := i.Bind(r, &req); err != nil {
		i.JSON(&w, i.A{2, ""})
		return
	}
	signature, err := models.SignPayload([]byte(requestedItemIds(r)))
	if err != nil {
		i.JSON(&w, i.A{2, ""})
		return
	}
	i.JSON(&w, i.A{0, signature})
}

// requestedItemIds returns the itemIDs value verbatim, whether it arrived in the
// form body or the query string. i.Bind has already parsed the form by the time
// this runs.
func requestedItemIds(r *http.Request) string {
	if r.Method == http.MethodGet {
		return r.URL.Query().Get("itemIDs")
	}
	if value := r.PostForm.Get("itemIDs"); value != "" {
		return value
	}
	return r.URL.Query().Get("itemIDs")
}
