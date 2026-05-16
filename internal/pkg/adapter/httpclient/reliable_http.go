package httpclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/sony/gobreaker/v2"
)

type Config struct {
	Timeout              time.Duration        `config:"timeout"`
	Delay                time.Duration        `config:"delay"`
	MaxDelay             time.Duration        `config:"max_delay"`
	DelayType            string               `config:"delay_type"`
	Retries              int                  `config:"retries"`
	RetriableStatusCodes []int                `config:"retriable_codes"`
	CircuitBreaker       CircuitBreakerConfig `config:"circuit_breaker"`
}

type CircuitBreakerConfig struct {
	MinRequiredCalls     int           `config:"min_required_calls"`
	MaxRequestsHalfOpen  int           `config:"max_requests_half_open_in_half_open"`
	ClosedStateInterval  time.Duration `config:"closed_state_interval"`
	OpenStateInterval    time.Duration `config:"open_state_interval"`
	SlidingWindowSize    time.Duration `config:"sliding_window_size"`
	FailureRateThreshold int           `config:"failure_rate_threshold"`
}

type ReliableHTTPClient struct {
	httpClient *http.Client
	cfg        Config
	cb         *gobreaker.CircuitBreaker[*http.Response]
}

func NewReliableHTTPClient(cfg Config) *ReliableHTTPClient {
	cb := gobreaker.NewCircuitBreaker[*http.Response](gobreaker.Settings{ //nolint:bodyclose// thats just declaration
		MaxRequests:  uint32(cfg.CircuitBreaker.MaxRequestsHalfOpen),
		Interval:     cfg.CircuitBreaker.ClosedStateInterval,
		Timeout:      cfg.CircuitBreaker.OpenStateInterval,
		BucketPeriod: cfg.CircuitBreaker.SlidingWindowSize,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < uint32(cfg.CircuitBreaker.MinRequiredCalls) {
				return false
			}
			failureRate := float64(counts.TotalFailures) / float64(counts.Requests) * 100 //nolint:mnd // we making percentage here

			return failureRate > float64(cfg.CircuitBreaker.FailureRateThreshold)
		},
	})

	return &ReliableHTTPClient{httpClient: &http.Client{}, cfg: cfg, cb: cb}
}

func (c *ReliableHTTPClient) Do(req *http.Request) (*http.Response, error) {
	ctx, _ := context.WithTimeout(req.Context(), c.cfg.Timeout) //nolint:govet // calling cancel kills functions that call this one and nothing works

	req = req.WithContext(ctx)

	var resp *http.Response
	var err error

	var delayType retry.DelayTypeFunc

	switch c.cfg.DelayType {
	case "constant":
		delayType = retry.FixedDelay
	case "exp":
		delayType = retry.BackOffDelay
	case "random":
		delayType = retry.RandomDelay
	default:
		delayType = retry.FixedDelay
	}

	err = retry.New(retry.Attempts(uint(c.cfg.Retries)), retry.DelayType(delayType),
		retry.Delay(c.cfg.Delay)).Do(
		func() error {
			retryReq := req.Clone(req.Context())
			resp, err = c.cb.Execute(func() (*http.Response, error) { //nolint:bodyclose // closed on error; on success caller owns the body
				r, doErr := c.httpClient.Do(retryReq)
				if doErr != nil {
					slog.Error("error doing request", "request", retryReq)
					if !errors.Is(doErr, context.DeadlineExceeded) {
						return nil, retry.Unrecoverable(doErr)
					}
					return nil, fmt.Errorf("requesting %s: %w", req.URL, doErr)
				}
				if r.StatusCode == http.StatusOK || r.StatusCode == http.StatusCreated {
					return r, nil
				}
				_ = r.Body.Close()
				if slices.Contains(c.cfg.RetriableStatusCodes, r.StatusCode) {
					slog.Error("error requesting with retryable status code", "statusCode", r.StatusCode)
					return nil, fmt.Errorf("unexpected status code: %d", r.StatusCode)
				}
				slog.Info("error requesting with non retryable status code", "statusCode", r.StatusCode)
				return nil, retry.Unrecoverable(errors.New("non-retriable status code" + strconv.Itoa(r.StatusCode)))
			})

			if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
				return retry.Unrecoverable(fmt.Errorf("circuit breaker: %w", err))
			}
			if err != nil {
				return fmt.Errorf("retry error: %w", err)
			}
			return nil
		})

	if err != nil {
		return nil, fmt.Errorf("requesting error %s: %w", req.URL, err)
	}

	return resp, nil
}
