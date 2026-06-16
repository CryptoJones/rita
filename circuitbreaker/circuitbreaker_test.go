package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeClock is a controllable clock for deterministic timeout tests.
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestBreaker(clk *fakeClock) *Breaker {
	return New(Options{
		FailureThreshold: 3,
		ResetTimeout:     10 * time.Second,
		HalfOpenMax:      2,
		Now:              clk.now,
	})
}

var errBoom = errors.New("boom")

func fail(context.Context) error { return errBoom }
func ok(context.Context) error   { return nil }

func TestTripsAfterThreshold(t *testing.T) {
	clk := &fakeClock{t: time.Unix(1000, 0)}
	b := newTestBreaker(clk)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if err := b.Execute(ctx, fail); !errors.Is(err, errBoom) {
			t.Fatalf("call %d: expected errBoom, got %v", i, err)
		}
	}
	if b.State() != StateClosed {
		t.Fatalf("expected still closed before threshold, got %s", b.State())
	}

	// third consecutive failure trips it
	if err := b.Execute(ctx, fail); !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom, got %v", err)
	}
	if b.State() != StateOpen {
		t.Fatalf("expected open after threshold, got %s", b.State())
	}

	// while open, calls fail fast without invoking fn
	called := false
	err := b.Execute(ctx, func(context.Context) error { called = true; return nil })
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
	if called {
		t.Fatal("fn should not be called while open")
	}
}

func TestSuccessResetsFailureCount(t *testing.T) {
	clk := &fakeClock{t: time.Unix(1000, 0)}
	b := newTestBreaker(clk)
	ctx := context.Background()

	_ = b.Execute(ctx, fail)
	_ = b.Execute(ctx, fail)
	_ = b.Execute(ctx, ok) // resets the consecutive-failure counter
	_ = b.Execute(ctx, fail)
	_ = b.Execute(ctx, fail)

	if b.State() != StateClosed {
		t.Fatalf("expected closed (counter was reset by success), got %s", b.State())
	}
}

func TestHalfOpenRecovery(t *testing.T) {
	clk := &fakeClock{t: time.Unix(1000, 0)}
	b := newTestBreaker(clk)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_ = b.Execute(ctx, fail)
	}
	if b.State() != StateOpen {
		t.Fatalf("precondition: expected open, got %s", b.State())
	}

	// not enough time elapsed -> still open
	clk.advance(5 * time.Second)
	if err := b.Execute(ctx, ok); !errors.Is(err, ErrOpen) {
		t.Fatalf("expected ErrOpen before reset timeout, got %v", err)
	}

	// reset timeout elapsed -> half-open, trial calls allowed
	clk.advance(10 * time.Second)
	if b.State() != StateHalfOpen {
		t.Fatalf("expected half-open after reset timeout, got %s", b.State())
	}

	// HalfOpenMax=2 successes required to close
	if err := b.Execute(ctx, ok); err != nil {
		t.Fatalf("unexpected error on first half-open success: %v", err)
	}
	if err := b.Execute(ctx, ok); err != nil {
		t.Fatalf("unexpected error on second half-open success: %v", err)
	}
	if b.State() != StateClosed {
		t.Fatalf("expected closed after enough half-open successes, got %s", b.State())
	}
}

func TestHalfOpenFailureReopens(t *testing.T) {
	clk := &fakeClock{t: time.Unix(1000, 0)}
	b := newTestBreaker(clk)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_ = b.Execute(ctx, fail)
	}
	clk.advance(10 * time.Second) // -> half-open

	if b.State() != StateHalfOpen {
		t.Fatalf("expected half-open, got %s", b.State())
	}
	// a single failure in half-open re-opens immediately
	if err := b.Execute(ctx, fail); !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom, got %v", err)
	}
	if b.State() != StateOpen {
		t.Fatalf("expected re-open after half-open failure, got %s", b.State())
	}
}
