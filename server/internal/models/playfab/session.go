package playfab

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand/v2"

	"github.com/luskaner/ageLANServer/server/internal"
)

var sessions *internal.SafeMap[string, string]

func generateId() string {
	bytes := make([]byte, 8)
	internal.WithRng(func(rand *rand.Rand) {
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
