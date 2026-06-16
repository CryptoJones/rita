// Package workerpool provides a tiny abstraction over the common "run N
// identical worker goroutines and wait for them" pattern.
//
// It replaces hand-managed sync.WaitGroup bookkeeping, where it is easy to
// mismatch Add/Done counts or to call a non-deferred Done that a panic or early
// return skips, leaking the group into a permanent Wait deadlock. Pool always
// calls Done via defer.
package workerpool

import "sync"

// Pool runs worker goroutines and waits for their completion. The zero value is
// ready to use. A Pool must not be copied after first use.
type Pool struct {
	wg sync.WaitGroup
}

// Go launches n workers, each running fn with its worker index (0..n-1). Done is
// deferred, so a worker that panics or returns early still releases the pool
// instead of deadlocking Wait. Go may be called multiple times; Wait blocks for
// every worker launched across all calls.
func (p *Pool) Go(n int, fn func(id int)) {
	p.wg.Add(n)
	for i := 0; i < n; i++ {
		go func(id int) {
			defer p.wg.Done()
			fn(id)
		}(i)
	}
}

// Wait blocks until all launched workers have returned.
func (p *Pool) Wait() {
	p.wg.Wait()
}
