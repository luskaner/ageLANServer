package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/luskaner/ageServer/server/internal/events"
)

// HandleEvents is GET /api/events: a server-sent event stream.
//
// It pushes a full snapshot on connect, on every event of the internal bus and
// every SnapshotSeconds, so a page only needs one connection and no polling
// logic. Browsers reconnect on their own and send Last-Event-ID; because every
// message is a complete snapshot, a missed event is irrelevant.
func (s *Server) HandleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	interval := time.Duration(s.opts.Config.SnapshotSeconds) * time.Second
	if interval <= 0 {
		interval = 2 * time.Second
	}

	bus, cancel := events.Subscribe()
	defer cancel()

	writeSnapshot := func() error {
		payload, err := json.Marshal(s.BuildSnapshot())
		if err != nil {
			return err
		}
		if _, err = fmt.Fprintf(w, "event: snapshot\ndata: %s\n\n", payload); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	if err := writeSnapshot(); err != nil {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case event, open := <-bus:
			if !open {
				return
			}
			// Forward the raw event immediately (the panel uses them to show a
			// toast) and mark the snapshot as stale.
			payload, err := json.Marshal(event)
			if err == nil {
				if _, err = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, event.Type, payload); err != nil {
					return
				}
				flusher.Flush()
			}
		case <-ticker.C:
			if err := writeSnapshot(); err != nil {
				return
			}
		case <-heartbeat.C:
			if _, err := fmt.Fprintf(w, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
