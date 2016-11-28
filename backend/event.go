package backend

import (
	"encoding/json"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Sirupsen/logrus"
	"net/http"
	"time"
)

func (b *Backend) EventStream(w http.ResponseWriter) error {
	notify := w.(http.CloseNotifier).CloseNotify()
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	for {
		select {
		case <-notify:
			logrus.Info("SSE Client closed")
			return nil
		case <-ticker.C:
			ev := b.sched.EventManager().Next()
			if ev != nil {
				WriteJSON(w, ev.(*types.Event))
			}

		}
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, message *types.Event) {
	f, _ := w.(http.Flusher)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	json.NewEncoder(w).Encode(message)
	f.Flush()
}
