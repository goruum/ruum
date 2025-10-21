package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/goruum/ruum/core"
)

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state    CircuitState
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{CircuitState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	config := DefaultCircuitBreakerConfig()

	if config.MaxRequests != 1 {
		t.Errorf("MaxRequests = %v, want 1", config.MaxRequests)
	}

	if config.Interval != 0 {
		t.Errorf("Interval = %v, want 0", config.Interval)
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", config.Timeout)
	}

	if config.ReadyToTrip == nil {
		t.Error("ReadyToTrip should not be nil")
	}

	if config.IsSuccessful == nil {
		t.Error("IsSuccessful should not be nil")
	}

	// Test ReadyToTrip default behavior
	counts := Counts{ConsecutiveFailures: 6}
	if !config.ReadyToTrip(counts) {
		t.Error("ReadyToTrip should return true for 6 consecutive failures")
	}

	counts = Counts{ConsecutiveFailures: 5}
	if config.ReadyToTrip(counts) {
		t.Error("ReadyToTrip should return false for 5 consecutive failures")
	}

	// Test IsSuccessful default behavior
	if !config.IsSuccessful(nil) {
		t.Error("IsSuccessful should return true for nil error")
	}

	if config.IsSuccessful(errors.New("error")) {
		t.Error("IsSuccessful should return false for non-nil error")
	}
}

func TestNewCircuitBreaker(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxRequests: 0, // Should default to 1
	}

	cb := newCircuitBreaker(config)

	if cb.config.MaxRequests != 1 {
		t.Errorf("MaxRequests should default to 1, got %v", cb.config.MaxRequests)
	}

	if cb.config.ReadyToTrip == nil {
		t.Error("ReadyToTrip should have default value")
	}

	if cb.config.IsSuccessful == nil {
		t.Error("IsSuccessful should have default value")
	}

	if cb.state != StateClosed {
		t.Errorf("Initial state should be Closed, got %v", cb.state)
	}
}

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := newCircuitBreaker(config)

	callCount := 0
	err := cb.Execute(func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("Execute() should succeed, got error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Function should be called once, got %d", callCount)
	}

	if cb.GetState() != StateClosed {
		t.Errorf("State should remain Closed, got %v", cb.GetState())
	}

	counts := cb.GetCounts()
	if counts.TotalSuccesses != 1 {
		t.Errorf("TotalSuccesses = %v, want 1", counts.TotalSuccesses)
	}
}

func TestCircuitBreaker_Execute_Failure(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.ReadyToTrip = func(counts Counts) bool {
		return counts.ConsecutiveFailures >= 3
	}
	cb := newCircuitBreaker(config)

	testErr := errors.New("test error")

	// First failure
	err := cb.Execute(func() error {
		return testErr
	})
	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}
	if cb.GetState() != StateClosed {
		t.Error("Should remain closed after 1 failure")
	}

	counts1 := cb.GetCounts()
	if counts1.ConsecutiveFailures != 1 {
		t.Errorf("After 1 failure: ConsecutiveFailures = %v, want 1", counts1.ConsecutiveFailures)
	}

	// Second failure
	_ = cb.Execute(func() error {
		return testErr
	})
	if cb.GetState() != StateClosed {
		t.Error("Should remain closed after 2 failures")
	}

	counts2 := cb.GetCounts()
	if counts2.ConsecutiveFailures != 2 {
		t.Errorf("After 2 failures: ConsecutiveFailures = %v, want 2", counts2.ConsecutiveFailures)
	}

	// Third failure - should open circuit
	_ = cb.Execute(func() error {
		return testErr
	})
	if cb.GetState() != StateOpen {
		t.Errorf("Should be open after 3 failures, got %v", cb.GetState())
	}

	// After state change to Open, counts are reset (new generation)
	// This is expected behavior in circuit breaker pattern
	counts3 := cb.GetCounts()
	if counts3.Requests != 0 {
		t.Errorf("Counts should be reset after state change, Requests = %v", counts3.Requests)
	}
}

func TestCircuitBreaker_OpenState_BlocksRequests(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.ReadyToTrip = func(counts Counts) bool {
		return counts.ConsecutiveFailures >= 1
	}
	cb := newCircuitBreaker(config)

	// Trigger circuit to open
	_ = cb.Execute(func() error {
		return errors.New("failure")
	})

	if cb.GetState() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Try to execute - should be blocked
	err := cb.Execute(func() error {
		t.Error("Function should not be called when circuit is open")
		return nil
	})

	if err == nil {
		t.Error("Execute should return error when circuit is open")
	}

	if err.Error() != "circuit breaker is open" {
		t.Errorf("Error message = %v, want 'circuit breaker is open'", err.Error())
	}
}

func TestCircuitBreaker_HalfOpen_Transition(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.Timeout = 50 * time.Millisecond
	config.MaxRequests = 2
	config.ReadyToTrip = func(counts Counts) bool {
		return counts.ConsecutiveFailures >= 1
	}
	cb := newCircuitBreaker(config)

	// Open the circuit
	_ = cb.Execute(func() error {
		return errors.New("failure")
	})

	if cb.GetState() != StateOpen {
		t.Fatal("Circuit should be open")
	}

	// Wait for timeout to transition to half-open
	time.Sleep(60 * time.Millisecond)

	// This should work now (half-open allows limited requests)
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("First request in half-open should succeed, got error: %v", err)
	}

	if cb.GetState() != StateHalfOpen {
		t.Errorf("State should be HalfOpen, got %v", cb.GetState())
	}

	// Second successful request should close the circuit
	err = cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Second request should succeed, got error: %v", err)
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Circuit should be closed after successful half-open requests, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpen_LimitRequests(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.Timeout = 50 * time.Millisecond
	config.MaxRequests = 1
	config.ReadyToTrip = func(counts Counts) bool {
		return counts.ConsecutiveFailures >= 1
	}
	cb := newCircuitBreaker(config)

	// Open the circuit
	_ = cb.Execute(func() error {
		return errors.New("failure")
	})

	// Wait for half-open
	time.Sleep(60 * time.Millisecond)

	// Start first request concurrently
	done := make(chan bool)
	go func() {
		_ = cb.Execute(func() error {
			time.Sleep(20 * time.Millisecond) // Simulate slow request
			return nil
		})
		done <- true
	}()

	// Give first request time to start
	time.Sleep(5 * time.Millisecond)

	// Second request should be rejected (MaxRequests = 1)
	err := cb.Execute(func() error {
		t.Error("Should not execute when half-open limit is reached")
		return nil
	})

	<-done // Wait for first request to complete

	if err == nil {
		t.Error("Should return error when half-open limit is reached")
	}

	if err != nil && err.Error() != "circuit breaker is half-open with too many requests" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestCircuitBreaker_HalfOpen_FailureReopens(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.Timeout = 50 * time.Millisecond
	config.MaxRequests = 1
	config.ReadyToTrip = func(counts Counts) bool {
		return counts.ConsecutiveFailures >= 1
	}
	cb := newCircuitBreaker(config)

	// Open the circuit
	_ = cb.Execute(func() error {
		return errors.New("failure")
	})

	// Wait for half-open
	time.Sleep(60 * time.Millisecond)

	// Fail in half-open state - should reopen circuit
	_ = cb.Execute(func() error {
		return errors.New("failure again")
	})

	if cb.GetState() != StateOpen {
		t.Errorf("Circuit should reopen after failure in half-open, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_OnStateChange_Callback(t *testing.T) {
	stateChanges := []string{}
	config := DefaultCircuitBreakerConfig()
	config.ReadyToTrip = func(counts Counts) bool {
		return counts.ConsecutiveFailures >= 1
	}
	config.OnStateChange = func(from, to CircuitState) {
		stateChanges = append(stateChanges, from.String()+" -> "+to.String())
	}
	cb := newCircuitBreaker(config)

	// Trigger state change to open
	_ = cb.Execute(func() error {
		return errors.New("failure")
	})

	if len(stateChanges) != 1 {
		t.Errorf("Expected 1 state change, got %d", len(stateChanges))
	}

	if stateChanges[0] != "closed -> open" {
		t.Errorf("Expected 'closed -> open', got %v", stateChanges[0])
	}
}

func TestCircuitBreaker_Interval_ResetsCount(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.Interval = 100 * time.Millisecond
	config.ReadyToTrip = func(counts Counts) bool {
		return counts.ConsecutiveFailures >= 5
	}
	cb := newCircuitBreaker(config)

	// Add some failures (but not enough to open)
	_ = cb.Execute(func() error {
		return errors.New("failure")
	})
	_ = cb.Execute(func() error {
		return errors.New("failure")
	})

	counts := cb.GetCounts()
	if counts.ConsecutiveFailures != 2 {
		t.Errorf("ConsecutiveFailures = %v, want 2", counts.ConsecutiveFailures)
	}

	// Wait for interval to pass
	time.Sleep(150 * time.Millisecond)

	// Make a request to trigger state check
	_ = cb.Execute(func() error {
		return nil
	})

	// Counts should be reset
	counts = cb.GetCounts()
	if counts.ConsecutiveFailures != 0 {
		t.Errorf("ConsecutiveFailures should be reset to 0, got %v", counts.ConsecutiveFailures)
	}
	if counts.ConsecutiveSuccesses != 1 {
		t.Errorf("ConsecutiveSuccesses = %v, want 1", counts.ConsecutiveSuccesses)
	}
}

func TestCircuitBreaker_Execute_Panic_Recovery(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := newCircuitBreaker(config)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Execute should propagate panic")
		}
	}()

	_ = cb.Execute(func() error {
		panic("test panic")
	})
}

func TestCircuitBreaker_Execute_Panic_RecordsFailure(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := newCircuitBreaker(config)

	func() {
		defer func() {
			_ = recover()
		}()
		_ = cb.Execute(func() error {
			panic("test panic")
		})
	}()

	counts := cb.GetCounts()
	if counts.TotalFailures != 1 {
		t.Errorf("Panic should be recorded as failure, TotalFailures = %v", counts.TotalFailures)
	}
}

func TestCircuitBreaker_Concurrent_Access(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.MaxRequests = 100
	cb := newCircuitBreaker(config)

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Concurrent successful requests
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cb.Execute(func() error {
				time.Sleep(time.Millisecond)
				return nil
			})
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successCount != 50 {
		t.Errorf("Expected 50 successful requests, got %d", successCount)
	}

	counts := cb.GetCounts()
	if counts.TotalSuccesses != 50 {
		t.Errorf("TotalSuccesses = %v, want 50", counts.TotalSuccesses)
	}
}

func TestCircuitBreakerMiddleware_Success(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	middleware := CircuitBreaker(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Middleware should succeed, got error: %v", err)
	}
}

func TestCircuitBreakerMiddleware_CircuitOpen(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.ReadyToTrip = func(counts Counts) bool {
		return counts.ConsecutiveFailures >= 1
	}
	middleware := CircuitBreaker(config)

	handler := func(ctx core.Context) error {
		return errors.New("handler error")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	// First request - should fail and open circuit
	_ = wrappedHandler(ctx)

	// Second request - should be blocked
	req2 := httptest.NewRequest("GET", "/test", nil)
	res2 := httptest.NewRecorder()
	ctx2 := core.NewContext(context.Background(), req2, res2, core.NewContainer())

	err := wrappedHandler(ctx2)
	if err == nil {
		t.Error("Should return error when circuit is open")
	}

	header := res2.Header().Get("X-Circuit-Breaker")
	if header != "open" {
		t.Errorf("X-Circuit-Breaker header = %v, want 'open'", header)
	}
}

func TestCircuitBreakerMiddleware_HalfOpenLimited(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.Timeout = 50 * time.Millisecond
	config.MaxRequests = 1
	config.ReadyToTrip = func(counts Counts) bool {
		return counts.ConsecutiveFailures >= 1
	}
	middleware := CircuitBreaker(config)

	handler := func(ctx core.Context) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	}

	wrappedHandler := middleware(handler)

	// Open circuit
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())
	_ = wrappedHandler(core.NewContext(context.Background(), req, res, core.NewContainer()))

	// Force fail to open circuit
	_ = middleware(func(ctx core.Context) error {
		return errors.New("fail")
	})(ctx)

	// Wait for half-open
	time.Sleep(60 * time.Millisecond)

	// Start first request (slow)
	go func() {
		req1 := httptest.NewRequest("GET", "/test", nil)
		res1 := httptest.NewRecorder()
		ctx1 := core.NewContext(context.Background(), req1, res1, core.NewContainer())
		_ = wrappedHandler(ctx1)
	}()

	time.Sleep(5 * time.Millisecond)

	// Try second request immediately - should be rejected
	req2 := httptest.NewRequest("GET", "/test", nil)
	res2 := httptest.NewRecorder()
	ctx2 := core.NewContext(context.Background(), req2, res2, core.NewContainer())

	err := wrappedHandler(ctx2)
	if err == nil {
		t.Error("Should return error when half-open limit is reached")
	}

	header := res2.Header().Get("X-Circuit-Breaker")
	if header != "half-open-limited" {
		t.Errorf("X-Circuit-Breaker header = %v, want 'half-open-limited'", header)
	}
}

