package mdatp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	thirtyDays     = 30 * 24 * time.Hour
	maxLookBehind  = thirtyDays - 1*time.Second
	maxInterval    = 24 * time.Hour
	datetimeFormat = "2006-01-02T15:04:05.99999Z"
)

// AlertWatchRequest .
type AlertWatchRequest struct {
	OutputSource   io.ReadWriteCloser
	IsOutputIndent bool

	HasStateSource bool
	StateSource    io.ReadWriteCloser

	Logger *logrus.Logger
	State  WatchState
}

// Watch retrieves alerts at regular intervals and write
// the results to the provided io.writer.
func (s *AlertService) Watch(ctx context.Context, req *AlertWatchRequest) error {
	req.Logger.Info("starting..")
	defer req.Logger.Info("stopped")
	if req.HasStateSource {
		req.Logger.Debug("using state source")
		defer req.State.Save(req.StateSource)
		if err := req.State.Load(req.StateSource); err != nil {
			req.Logger.Warnf("could not load state: %v", err)
		}
	}

	var wg sync.WaitGroup

	alertBufSize := 1024
	alertCh := make(chan Alert, alertBufSize)

	tickerInterval := 5 * time.Second
	ticker := time.NewTicker(tickerInterval)

	encodeDoneCh := make(chan struct{})

	cancelCtx, cancel := context.WithCancel(ctx)

	wg.Add(1)
	go func() {
		defer cancel()
		defer close(alertCh)
		defer wg.Done()

		var lock uint64

		for {
			select {
			case <-encodeDoneCh:
				return
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				req.Logger.Debug("triggered")
				if !atomic.CompareAndSwapUint64(&lock, 0, 1) {
					req.Logger.Info("busy querying..")
					continue
				}

				go func() {
					defer atomic.StoreUint64(&lock, 0)

					var err error
					var start time.Time
					end := now
					for {
						start, err = req.State.GetLastFetchTime()
						if err != nil {
							req.Logger.Debugf("could not get lastFetchTime: %v", err)
						}
						if !start.Before(now) {
							req.Logger.Debug("we are done looping")
							return
						}

						if start.IsZero() {
							// ? TODO: firstFetchLookBehind value in config maybe ?
							start = end.Add(-tickerInterval)
						}
						if start.Before(now.Add(-maxLookBehind)) {
							end = end.Add(-(maxLookBehind))
							req.Logger.Debugf("start is older than allowed value(%v): %v", maxLookBehind.String(), start.String())
						}
						interval := end.Sub(start)
						if interval > maxInterval {
							end = start.Add(maxInterval)
							req.Logger.Debugf("interval is greater than allowed value(%v): %v", maxInterval.String(), interval.String())
						}

						oDataIntervalQueryStr := "alertCreationTime gt %v and alertCreationTime le %v"
						oDataIntervalQuery := fmt.Sprintf(oDataIntervalQueryStr, start.UTC().Format(datetimeFormat), end.UTC().Format(datetimeFormat))
						req.Logger.Debugf("ODATA filter query: %v", oDataIntervalQuery)

						resp, alert, err := s.client.Alert.List(cancelCtx, oDataIntervalQuery)
						if err != nil {
							if !errors.Is(err, context.Canceled) {
								req.Logger.Errorf("request error: %v", err)
							}
							return
						}
						if resp.APIError != nil {
							req.Logger.Errorf("api error: %+v", resp.APIError)
							return
						}
						req.Logger.Debug("query succesfull")
						for _, a := range alert.Value {
							select {
							case alertCh <- a:
							case <-cancelCtx.Done():
								return
							}
						}
						req.State.SetLastFetchTime(end)
					}
				}()

			}
		}
	}()

	enc := json.NewEncoder(req.OutputSource)
	if req.IsOutputIndent {
		enc.SetIndent("", "\t")
	}

	wg.Add(1)
	go func() {
		defer close(encodeDoneCh)
		defer wg.Done()
		for alert := range alertCh {
			if err := enc.Encode(&alert); err != nil {
				req.Logger.Error(err)
				return
			}
		}
	}()

	req.Logger.Info("started")
	wg.Wait()
	return nil
}

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
