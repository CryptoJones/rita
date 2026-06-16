package workerpool

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunsAllWorkers(t *testing.T) {
	var p Pool
	var count atomic.Int64
	p.Go(10, func(int) { count.Add(1) })
	p.Wait()
	if got := count.Load(); got != 10 {
		t.Fatalf("expected 10 workers to run, got %d", got)
	}
}

func TestMultipleGoAccumulate(t *testing.T) {
	var p Pool
	var count atomic.Int64
	p.Go(3, func(int) { count.Add(1) })
	p.Go(4, func(int) { count.Add(1) })
	p.Wait()
	if got := count.Load(); got != 7 {
		t.Fatalf("expected 7 workers across two Go calls, got %d", got)
	}
}

func TestWaitBlocksUntilWorkersFinish(t *testing.T) {
	var p Pool
	var done atomic.Bool
	release := make(chan struct{})

	p.Go(5, func(int) {
		<-release // block until the test releases us
		done.Store(true)
	})

	// workers are blocked; Wait must not have returned, so done stays false
	if done.Load() {
		t.Fatal("workers finished before being released")
	}
	close(release)
	p.Wait()
	if !done.Load() {
		t.Fatal("Wait returned before workers finished")
	}
}

func TestWorkerIDsAreUnique(t *testing.T) {
	var p Pool
	const n = 16
	var mu sync.Mutex
	seen := make(map[int]bool)
	p.Go(n, func(id int) {
		mu.Lock()
		seen[id] = true
		mu.Unlock()
	})
	p.Wait()
	if len(seen) != n {
		t.Fatalf("expected ids 0..%d, got %d distinct ids", n-1, len(seen))
	}
	for i := 0; i < n; i++ {
		if !seen[i] {
			t.Errorf("missing worker id %d", i)
		}
	}
}

// TestWaitOnEmptyPoolReturns ensures Wait on a pool with no workers is a no-op.
func TestWaitOnEmptyPoolReturns(t *testing.T) {
	var p Pool
	doneCh := make(chan struct{})
	go func() { p.Wait(); close(doneCh) }()
	select {
	case <-doneCh:
	case <-time.After(time.Second):
		t.Fatal("Wait on empty pool blocked")
	}
}
