package mdatp

import (
	"encoding/json"
	"io"
	"time"
)

// WatchState .
type WatchState interface {
	SetLastFetchTime(time.Time) error
	GetLastFetchTime() (time.Time, error)
	Save(io.ReadWriteCloser) error
	Load(io.ReadWriteCloser) error
}

// WatchStateJSON .
type WatchStateJSON struct {
	lastFetchTime time.Time
}

// NewWatchStateJSON returns a WatchState using the provided
// source as persistence mecanism.
func NewWatchStateJSON() *WatchStateJSON {
	return &WatchStateJSON{time.Time{}}
}

// SetLastFetchTime implements the WatchState interface.
func (s *WatchStateJSON) SetLastFetchTime(t time.Time) error {
	s.lastFetchTime = t
	return nil
}

// GetLastFetchTime implements the WatchState interface.
func (s *WatchStateJSON) GetLastFetchTime() (time.Time, error) {
	return s.lastFetchTime, nil
}

// Save implements the WatchState interface.
func (s *WatchStateJSON) Save(rwc io.ReadWriteCloser) error {
	return json.NewEncoder(rwc).Encode(s)
}

// Load implements the WatchState interface.
func (s *WatchStateJSON) Load(rwc io.ReadWriteCloser) error {
	return json.NewDecoder(rwc).Decode(s)
}
