package playfab

import (
	"encoding/binary"
	"encoding/hex"
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal"
)

var sessions *internal.SafeMap[string, string]

func generateId() string {
	bytes := make([]byte, 8)
	internal.WithRng(func(rand *internal.RandReader) {
		binary.BigEndian.PutUint64(bytes, rand.Uint64())
	})
	return hex.EncodeToString(bytes)
}

func init() {
	sessions = internal.NewSafeMap[string, string]()
}

func Id(session string) (playfabId string, ok bool) {
	return sessions.Load(session)
}

func AddSession(session string) string {
	id := generateId()
	sessions.Store(session, id, nil)
	return id
}

func Session(r *http.Request) string {
	return r.Header.Get("X-Entitytoken")
}
