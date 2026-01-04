package login

import (
	"encoding/json"
	"fmt"
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func returnData(w http.ResponseWriter, messageId uint, messages []i.A) {
	j, _ := json.Marshal(i.A{messages})
	i.RawJSON(&w, []byte(fmt.Sprintf(`%d,%s`, messageId, j)))
}

func returnError(w http.ResponseWriter) {
	returnData(w, 0, []i.A{})
}

type readSessionRequest struct {
	Ack uint `schema:"ack"`
}

func ReadSession(w http.ResponseWriter, r *http.Request) {
	var req readSessionRequest
	err := i.Bind(r, &req)
	if err != nil {
		returnError(w)
		return
	}
	sess := models.SessionOrPanic(r)
	messageId, messages := sess.WaitForMessages(req.Ack)
	returnData(w, messageId, messages)
}
