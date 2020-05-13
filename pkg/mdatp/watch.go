package mdatp

import (
	"encoding/json"
	"io"
	"time"
)

// ReadWriteCloserMaker is used to save/load state.
type ReadWriteCloserMaker func() (io.ReadWriteCloser, error)

// WatchState .
type WatchState interface {
	SetLastFetchTime(time.Time) error
	GetLastFetchTime() (time.Time, error)
	Save(ReadWriteCloserMaker) error
	Load(ReadWriteCloserMaker) error
}

// WatchStateJSON .
type WatchStateJSON struct {
	LastFetchTime time.Time `json:"lastFetchTime"`
}

// NewWatchStateJSON returns a WatchState using the provided
// source as persistence mecanism.
func NewWatchStateJSON() *WatchStateJSON {
	return &WatchStateJSON{time.Time{}}
}

// SetLastFetchTime implements the WatchState interface.
func (s *WatchStateJSON) SetLastFetchTime(t time.Time) error {
	s.LastFetchTime = t
	return nil
}

// GetLastFetchTime implements the WatchState interface.
func (s *WatchStateJSON) GetLastFetchTime() (time.Time, error) {
	return s.LastFetchTime, nil
}

// Save implements the WatchState interface.
func (s *WatchStateJSON) Save(rwcMaker ReadWriteCloserMaker) error {
	rwc, err := rwcMaker()
	if err != nil {
		return err
	}
	defer rwc.Close()
	return json.NewEncoder(rwc).Encode(s)
}

// Load implements the WatchState interface.
func (s *WatchStateJSON) Load(rwcMaker ReadWriteCloserMaker) error {
	rwc, err := rwcMaker()
	if err != nil {
		return err
	}
	defer rwc.Close()
	return json.NewDecoder(rwc).Decode(s)
}
