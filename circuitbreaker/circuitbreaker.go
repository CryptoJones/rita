// Package circuitbreaker provides a small, dependency-free circuit breaker for
// guarding external/database operations so that a failing dependency fails fast
// instead of cascading (every caller piling up on a dead ClickHouse, a
// timing-out threat-intel feed, etc.).
//
// It implements the standard three-state model:
//
//	Closed   - calls pass through; consecutive failures are counted.
//	Open      - calls fail immediately with ErrOpen until ResetTimeout elapses.
//	HalfOpen  - a limited number of trial calls are allowed; enough successes
//	            close the breaker, any failure re-opens it.
//
// The breaker is safe for concurrent use.
package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State is the breaker's current state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
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

// ErrOpen is returned by Execute when the breaker is open and rejecting calls.
var ErrOpen = errors.New("circuit breaker is open")

// Options configures a Breaker. Zero values are replaced with sane defaults.
type Options struct {
	// FailureThreshold is the number of consecutive failures that trips the
	// breaker from Closed to Open. Default 5.
	FailureThreshold int
	// ResetTimeout is how long the breaker stays Open before allowing trial
	// calls (transitioning to HalfOpen). Default 30s.
	ResetTimeout time.Duration
	// HalfOpenMax is the number of consecutive successes required in HalfOpen to
	// close the breaker. Default 1.
	HalfOpenMax int
	// Now is an injectable clock for deterministic tests. Defaults to time.Now.
	Now func() time.Time
}

// Breaker is a concurrency-safe circuit breaker.
type Breaker struct {
	mu                sync.Mutex
	opts              Options
	state             State
	failures          int
	halfOpenSuccesses int
	openedAt          time.Time
}

// New returns a Breaker configured with opts (defaults applied for zero fields).
func New(opts Options) *Breaker {
	if opts.FailureThreshold <= 0 {
		opts.FailureThreshold = 5
	}
	if opts.ResetTimeout <= 0 {
		opts.ResetTimeout = 30 * time.Second
	}
	if opts.HalfOpenMax <= 0 {
		opts.HalfOpenMax = 1
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	return &Breaker{opts: opts, state: StateClosed}
}

// State returns the current breaker state (accounting for an elapsed reset
// timeout that would move it from Open to HalfOpen).
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeHalfOpen()
	return b.state
}

// Execute runs fn if the breaker allows it, recording the outcome. It returns
// ErrOpen without calling fn when the breaker is open.
func (b *Breaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	if err := b.beforeCall(); err != nil {
		return err
	}

	err := fn(ctx)

	b.afterCall(err)
	return err
}

// beforeCall checks whether a call is permitted and advances Open->HalfOpen if
// the reset timeout has elapsed.
func (b *Breaker) beforeCall() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.maybeHalfOpen()
	if b.state == StateOpen {
		return ErrOpen
	}
	return nil
}

// afterCall records the result of a permitted call and transitions state.
func (b *Breaker) afterCall(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		// any failure (re)opens the breaker from HalfOpen, or trips it from
		// Closed once the threshold is reached.
		b.failures++
		b.halfOpenSuccesses = 0
		if b.state == StateHalfOpen || b.failures >= b.opts.FailureThreshold {
			b.trip()
		}
		return
	}

	switch b.state {
	case StateHalfOpen:
		b.halfOpenSuccesses++
		if b.halfOpenSuccesses >= b.opts.HalfOpenMax {
			b.reset()
		}
	default:
		b.reset()
	}
}

// maybeHalfOpen moves an Open breaker to HalfOpen once ResetTimeout has elapsed.
// Caller must hold b.mu.
func (b *Breaker) maybeHalfOpen() {
	if b.state == StateOpen && b.opts.Now().Sub(b.openedAt) >= b.opts.ResetTimeout {
		b.state = StateHalfOpen
		b.halfOpenSuccesses = 0
	}
}

// trip opens the breaker. Caller must hold b.mu.
func (b *Breaker) trip() {
	b.state = StateOpen
	b.openedAt = b.opts.Now()
	b.halfOpenSuccesses = 0
}

// reset closes the breaker and clears counters. Caller must hold b.mu.
func (b *Breaker) reset() {
	b.state = StateClosed
	b.failures = 0
	b.halfOpenSuccesses = 0
}
