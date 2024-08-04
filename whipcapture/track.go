package server

import (
	"log/slog"

	"github.com/puzpuzpuz/xsync/v3"
)

type track struct {
	subscribers *xsync.MapOf[string, SampleWriter]
	kind        string
	logger      *slog.Logger
}

func newTrack(kind string) *track {
	return &track{
		subscribers: xsync.NewMapOf[string, SampleWriter](),
		kind:        kind,
		logger:      slog.Default().WithGroup("TRACK"),
	}
}

func (t *track) Subscribe(id string, sw SampleWriter) {
	t.subscribers.Store(id, sw)
}

func (t *track) Unsubscribe(id string) {
	t.subscribers.Delete(id)
}

func (t *track) WriteSample(s Sample) error {
	t.subscribers.Range(func(key string, sub SampleWriter) bool {
		if err := sub.WriteSample(s); err != nil {
			t.subscribers.Delete(key)
		}
		return true
	})
	return nil
}

func (t *track) Kind() string {
	return t.kind
}
