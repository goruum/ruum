package middleware

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goruum/ruum/core"
)

// CircuitState represents the circuit breaker state
type CircuitState int

const (
	// StateClosed allows all requests
	StateClosed CircuitState = iota
	// StateOpen blocks all requests
	StateOpen
	// StateHalfOpen allows limited requests to test recovery
	StateHalfOpen
)

// String returns the string representation of the state
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	// MaxRequests is the maximum number of requests allowed in half-open state
	MaxRequests uint32
	// Interval is the cyclic period in closed state to clear internal counts
	Interval time.Duration
	// Timeout is the period of open state before transitioning to half-open
	Timeout time.Duration
	// ReadyToTrip is called with a copy of Counts when in closed state.
	// If it returns true, the CircuitBreaker will be placed into open state.
	ReadyToTrip func(counts Counts) bool
	// OnStateChange is called whenever the state changes
	OnStateChange func(from, to CircuitState)
	// IsSuccessful determines if a response should be considered successful
	IsSuccessful func(err error) bool
}

// Counts holds the numbers of requests and their successes/failures
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// DefaultCircuitBreakerConfig returns a default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxRequests: 1,
		Interval:    time.Duration(0), // No automatic reset
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
		OnStateChange: nil,
		IsSuccessful: func(err error) bool {
			return err == nil
		},
	}
}

type circuitBreaker struct {
	config     CircuitBreakerConfig
	mu         sync.RWMutex
	state      CircuitState
	counts     Counts
	expiry     time.Time // When the current state expires
	generation uint64    // Incremented when state changes to Closed or HalfOpen
}

func newCircuitBreaker(config CircuitBreakerConfig) *circuitBreaker {
	// Set default values if not provided
	if config.MaxRequests == 0 {
		config.MaxRequests = 1
	}
	if config.ReadyToTrip == nil {
		config.ReadyToTrip = func(counts Counts) bool {
			return counts.ConsecutiveFailures > 5
		}
	}
	if config.IsSuccessful == nil {
		config.IsSuccessful = func(err error) bool {
			return err == nil
		}
	}

	cb := &circuitBreaker{
		config: config,
		state:  StateClosed,
	}
	cb.toNewGeneration(time.Now())
	return cb
}

// Execute runs the given function if the circuit breaker allows it
func (cb *circuitBreaker) Execute(fn func() error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	result := fn()
	cb.afterRequest(generation, cb.config.IsSuccessful(result))
	return result
}

func (cb *circuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, errors.New("circuit breaker is open")
	}

	if state == StateHalfOpen && cb.counts.Requests >= cb.config.MaxRequests {
		return generation, errors.New("circuit breaker is half-open with too many requests")
	}

	cb.counts.Requests++
	return generation, nil
}

func (cb *circuitBreaker) afterRequest(before uint64, success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	// Don't skip recording if generation changed - that would lose failure tracking
	_ = before
	_ = generation

	if success {
		cb.onSuccess(state)
	} else {
		cb.onFailure(state)
	}
}

func (cb *circuitBreaker) onSuccess(state CircuitState) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0

	if state == StateHalfOpen && cb.counts.ConsecutiveSuccesses >= cb.config.MaxRequests {
		cb.setState(StateClosed)
	}
}

func (cb *circuitBreaker) onFailure(state CircuitState) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0

	if cb.config.ReadyToTrip(cb.counts) {
		cb.setState(StateOpen)
	}
}

func (cb *circuitBreaker) currentState(now time.Time) (CircuitState, uint64) {
	switch cb.state {
	case StateClosed:
		if cb.config.Interval > 0 && !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.setState(StateHalfOpen)
		}
	}
	return cb.state, cb.getGeneration()
}

func (cb *circuitBreaker) setState(state CircuitState) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(time.Now())

	if cb.config.OnStateChange != nil {
		cb.config.OnStateChange(prev, state)
	}
}

func (cb *circuitBreaker) toNewGeneration(now time.Time) {
	atomic.AddUint64(&cb.generation, 1)
	cb.counts = Counts{} // Reset counts for new generation

	switch cb.state {
	case StateClosed:
		if cb.config.Interval > 0 {
			cb.expiry = now.Add(cb.config.Interval)
		} else {
			cb.expiry = time.Time{} // No expiry if interval is 0
		}
	case StateOpen:
		cb.expiry = now.Add(cb.config.Timeout)
	case StateHalfOpen:
		cb.expiry = time.Time{} // Half-open has no fixed expiry, relies on MaxRequests
	}
}

func (cb *circuitBreaker) getGeneration() uint64 {
	return atomic.LoadUint64(&cb.generation)
}

// GetState returns the current state of the circuit breaker
func (cb *circuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetCounts returns a copy of the current counts
func (cb *circuitBreaker) GetCounts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.counts
}

// CircuitBreaker creates a circuit breaker middleware
func CircuitBreaker(config CircuitBreakerConfig) core.MiddlewareFunc {
	cb := newCircuitBreaker(config)

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) error {
			err := cb.Execute(func() error {
				return next(ctx)
			})

			if err != nil && err.Error() == "circuit breaker is open" {
				ctx.SetHeader("X-Circuit-Breaker", "open")
				return core.NewHTTPException(503, "Service temporarily unavailable")
			}

			if err != nil && err.Error() == "circuit breaker is half-open with too many requests" {
				ctx.SetHeader("X-Circuit-Breaker", "half-open-limited")
				return core.NewHTTPException(503, "Service temporarily unavailable")
			}

			return err
		}
	}
}
