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
)

var (
	oneDay     = 24 * time.Hour
	thirtyDays = 30 * oneDay
)

var (
	// minTickerInterval is a lower bound constraint so that we
	// do not exceed the Microsoft quota limitation:
	// 1500 query/hour ~= 0.42 query/second ~= 1 query/2.38 second.
	minTickerInterval = 3 * time.Second
	// maxTickerInterval is an upper bound constraint.
	maxTickerInterval = oneDay

	// minAlertInterval is a lower bound constraint.
	minAlertInterval = 1 * time.Second
	// maxAlertInterval is an upper bound constraint.
	maxAlertInterval = maxTickerInterval

	// maxLookBehind is a hard requirement from microsoft for the alertCreationTime field.
	maxLookBehind = thirtyDays - 1*time.Second
)

var (
	// defaultTickerInterval defines the ticker interval duration
	// to trigger a query to the API.
	defaultTickerInterval = minTickerInterval
	// defaultMaxAlertInterval defines the max interval allowed
	// for the alertCreationTime field in the OData filter query.
	defaultMaxAlertInterval = maxAlertInterval
)

var (
	// alertChSize defines the channel buffer size
	// when sending alerts to the encoder goroutine.
	alertChSize = 1024
)

// AlertWatchRequest defines attributes required by the Watch method.
type AlertWatchRequest struct {
	OutputSource   io.ReadWriteCloser
	IsOutputIndent bool

	State            WatchState
	StateSourceMaker ReadWriteCloserMaker
	HasStateSource   bool

	QueryInterval    int
	QueryMaxInterval int
}

// Watch retrieves alerts at regular intervals and writes
// the results to the provided OutputSource.
//
// Two goroutines are started, one to query and one to encode results.
// The query goroutine, for each tick, and if not already running,
// a query to the Alert endpoint is made. Alerts retrieved are sent
// to an alert channel to be encoded by the encoding goroutine.
//
// An error is returned if request attribute validation fails.
func (s *AlertService) Watch(ctx context.Context, req *AlertWatchRequest) error {
	tickerInterval := time.Duration(req.QueryInterval) * time.Second
	if tickerInterval == 0 {
		tickerInterval = defaultTickerInterval
	}
	if tickerInterval < minTickerInterval {
		return fmt.Errorf("tickerInterval is below the minimum allowed(%v): %v", minTickerInterval.String(), tickerInterval.String())
	}
	if tickerInterval > maxTickerInterval {
		return fmt.Errorf("tickerInterval is above the maxmimum allowed(%v): %v", maxTickerInterval.String(), tickerInterval.String())
	}

	maxInterval := time.Duration(req.QueryMaxInterval) * time.Minute
	if maxInterval == 0 {
		maxInterval = defaultMaxAlertInterval
	}
	if maxInterval < minAlertInterval {
		return fmt.Errorf("maxInterval is below the minimum allowed(%v): %v", minAlertInterval.String(), tickerInterval.String())
	}
	if tickerInterval > maxAlertInterval {
		return fmt.Errorf("maxInterval is above the maxmimum allowed(%v): %v", maxAlertInterval.String(), tickerInterval.String())
	}

	if req.HasStateSource {
		s.client.logger.Info("using state source")
		defer req.State.Save(req.StateSourceMaker)

		if err := req.State.Load(req.StateSourceMaker); err != nil {
			if err != io.EOF {
				s.client.logger.Warnf("could not load state: %v", err)
			} else {
				s.client.logger.Info("state is empty")
			}
		}
	}

	encoder := json.NewEncoder(req.OutputSource)
	if req.IsOutputIndent {
		encoder.SetIndent("", "\t")
	}

	encodeDoneCh := make(chan struct{})
	alertCh := make(chan Alert, alertChSize)

	var wg sync.WaitGroup

	queryFunc := func(ctx context.Context, lock *uint64, triggered time.Time) {
		defer atomic.StoreUint64(lock, 0)

		var err error
		var start time.Time
		end := triggered
		for {
			s.client.logger.Debug("retrieving lastFetchTime")
			start, err = req.State.GetLastFetchTime()
			if err != nil {
				s.client.logger.Debugf("could not get lastFetchTime from state: %v", err)
			}

			// first run without state, set start to maxinterval.
			if start.IsZero() {
				start = end.Add(-maxInterval)
			}

			interval := end.Sub(start)
			if interval == 0 {
				break
			}

			// we have looped, move end forward
			if end.Before(triggered) {
				end.Add(maxInterval)
			}
			// end has been moved farther than now(), set end to now()
			if end.After(triggered) {
				end = triggered
			}
			// validate max lookbehind and adjust start
			if interval > maxLookBehind {
				start = end.Add(-maxLookBehind)
			}
			// validate max interval and adjust end
			if interval > maxInterval {
				end = start.Add(maxInterval)
			}

			oDataIntervalQuery := makeIntervalOdataQuery("alertCreationTime", start, end)
			s.client.logger.Debugf("ODATA filter query: %v", oDataIntervalQuery)

			resp, alert, err := s.client.Alert.List(ctx, oDataIntervalQuery)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					s.client.logger.Errorf("request error: %v", err)
				}
				return
			}
			if resp.APIError != nil {
				s.client.logger.Errorf("api error: %+v", resp.APIError)
				return
			}
			s.client.logger.Debugf("query succesfull. Retrieved %d alerts.", len(alert.Value))
			for _, a := range alert.Value {
				select {
				case alertCh <- a:
				case <-ctx.Done():
					return
				}
			}
			s.client.logger.Debug("saving lastFetchTime")
			req.State.SetLastFetchTime(end)
		}
	}

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(tickerInterval)
		cancelCtx, cancel := context.WithCancel(ctx)

		defer func() {
			ticker.Stop()
			cancel()
			close(alertCh)
			wg.Done()
		}()

		var lock uint64
		go queryFunc(cancelCtx, &lock, time.Now())

		for {
			select {
			case <-encodeDoneCh:
				return
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				s.client.logger.Debug("triggered")
				if !atomic.CompareAndSwapUint64(&lock, 0, 1) {
					s.client.logger.Debug("busy..")
					continue
				}
				go queryFunc(cancelCtx, &lock, now)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer func() {
			close(encodeDoneCh)
			wg.Done()
		}()

		for alert := range alertCh {
			if err := encoder.Encode(&alert); err != nil {
				s.client.logger.Error(err)
				return
			}
		}
	}()

	s.client.logger.Info("started")
	wg.Wait()
	s.client.logger.Info("stopped")
	return nil
}
