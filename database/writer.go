package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/activecm/rita/v5/circuitbreaker"
	"github.com/activecm/rita/v5/config"
	"github.com/activecm/rita/v5/metrics"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

// batchBufferPool reuses the per-worker batch buffers across the many writer
// goroutines created during an import (one writer per log type, each with
// several workers), avoiding a fresh []Data allocation for every worker. Buffers
// are pooled by pointer so that returning one to the pool doesn't itself
// allocate a boxed slice header (staticcheck SA6002).
var batchBufferPool = sync.Pool{
	New: func() any { s := make([]Data, 0); return &s },
}

//go:generate go run go.uber.org/mock/mockgen -source=writer.go -destination=mock_database_test.go -package=database

type (
	Data any

	// Interface to allow creating a BulkWriter from a DB or a serverConn
	Database interface {
		getConn() driver.Conn
		GetContext() context.Context
		QueryParameters(clickhouse.Parameters) context.Context
		// Breaker returns the circuit breaker guarding writes to this
		// connection, or nil if none is configured (writes then proceed
		// unguarded). Sharing one breaker per connection lets a struggling
		// ClickHouse fail fast for all writers at once.
		Breaker() *circuitbreaker.Breaker
	}

	BulkWriter struct {
		db           Database
		conf         *config.Config
		WriteChannel chan Data
		ProgChannel  chan int
		WriteWg      *errgroup.Group // wait for writing to finish
		writerName   string          // used in error reporting
		batchSize    int
		query        string
		limiter      *rate.Limiter
		withProgress bool
		database     string
		closed       bool
		ctx          context.Context
		numWorkers   int
		batches      []int
		mu           sync.Mutex
		cond         *sync.Cond
	}
)

// NewBulkWriter creates a new writer object to write output data to collections
func NewBulkWriter(db Database, conf *config.Config, numWorkers int, database string, writerName string, query string, limiter *rate.Limiter, withProgress bool) *BulkWriter {

	analysisErrGroup, ctx := errgroup.WithContext(context.Background())
	writer := &BulkWriter{
		db:           db,
		conf:         conf,
		database:     database,
		WriteChannel: make(chan Data),
		ProgChannel:  make(chan int),
		WriteWg:      analysisErrGroup,
		writerName:   writerName,
		batchSize:    int(conf.RITA.BatchSize),
		query:        query,
		limiter:      limiter,
		withProgress: withProgress,
		numWorkers:   numWorkers,
		ctx:          ctx,
		batches:      make([]int, numWorkers), // keeps track of the batch count for each worker
	}
	writer.cond = sync.NewCond(&writer.mu)
	return writer
}

// shouldReadData returns whether or not the thread with the passed in ID should read data from the write channel
func (w *BulkWriter) shouldReadData(id int, empty bool) bool {
	if w.numWorkers == 1 {
		return true
	}

	var numInProgress int
	for i, b := range w.batches {
		if i != id {
			// batch is in progress if it has at least 1 item, but less than the batch size
			if b > 0 && b < w.batchSize {
				numInProgress++
			}
		}
	}
	// we don't want a worker that's not currently in progress to read the rest of the items from the channel after it's closed
	// because then the leftover data will get distributed between all of the workers, making 5 or so tiny batches, which is really bad
	if w.closed {
		// allow any worker to pass through the cond wait if the channel is empty
		if empty {
			return true
		}
		// if the channel isn't empty yet, allow any in progress workers to keep going, or a new one if none are processing
		return w.batches[id] > 0 || numInProgress == 0
	}

	// a worker should start reading if there are no other workers currently reading in data
	// or keep reading if it's already in progress
	return numInProgress == 0 || w.batches[id] > 0
}

// Close waits for the write threads to finish. It returns the first error
// encountered by any writer worker (if any) rather than terminating the process,
// leaving it to the caller to decide how a write failure should be handled.
func (w *BulkWriter) Close() error {
	// tell workers that no more data will be sent on this channel
	close(w.WriteChannel)
	// mark the channel as closed
	w.closed = true
	// notify workers that the channel is closed
	w.cond.Broadcast()
	// wait for the errgroup
	err := w.WriteWg.Wait()

	close(w.ProgChannel)

	if err != nil {
		metrics.WriteErrorsTotal.Inc()
		return fmt.Errorf("error writing to database for %q: %w", w.writerName, err)
	}
	return nil
}

// sendBatch sends a prepared batch, routed through the connection's circuit
// breaker when one is configured so a struggling ClickHouse fails fast for every
// writer at once instead of each piling up on dial/timeout waits.
func (w *BulkWriter) sendBatch(batch driver.Batch) error {
	if br := w.db.Breaker(); br != nil {
		return br.Execute(w.ctx, func(context.Context) error { return batch.Send() })
	}
	return batch.Send()
}

// Start kicks off a new write thread
func (w *BulkWriter) Start(id int) {

	w.WriteWg.Go(func() error {
		conn := w.db.getConn()

		chCtx := w.db.QueryParameters(clickhouse.Parameters{
			"database": w.database,
		})

		batchCount := 0

		// borrow a batch buffer from the shared pool and return it (cleared) when
		// this worker exits, so its backing array can be reused by another worker.
		bufPtr := batchBufferPool.Get().(*[]Data)
		items := (*bufPtr)[:0]
		defer func() {
			clear(items[:cap(items)])
			*bufPtr = items[:0]
			batchBufferPool.Put(bufPtr)
		}()

		// loop over input channel
		for {

			w.mu.Lock()
			// check to see if this thread should take in data
			for !w.shouldReadData(id, len(w.WriteChannel) == 0) {
				// wait for other threads to process data if it isn't supposed to read in data yet
				w.cond.Wait()
			}

			// check if any other workers errored out and made the context finish
			select {
			case <-w.ctx.Done():
				return w.ctx.Err()
			default:
			}

			// attempt to read data from the channel
			change, ok := <-w.WriteChannel

			// if the channel is closed, unlock the mutex and break out of the loop
			if !ok {
				w.mu.Unlock()
				break // Exit if the channel is closed
			}
			// increment batch count
			w.batches[id]++
			batchCount++
			// unlock mutex
			w.mu.Unlock()

			// add this data to the batch buffer
			items = append(items, change)

			// if batch size limit reached, write out batch of records
			if batchCount >= w.batchSize {
				// alert other workers that this worker is sending the batch so that
				// a free worker can be allowed to start making a new batch
				w.cond.Broadcast()

				// initialize batch
				batch, err := conn.PrepareBatch(chCtx, w.query)
				if err != nil {
					return fmt.Errorf("preparing batch for %q (batch_size %d): %w", w.writerName, w.batches[id], err)
				}

				// add each item in batch to this batch
				for _, item := range items {
					if err := batch.AppendStruct(item); err != nil {
						return fmt.Errorf("appending row to batch for %q (batch_size %d): %w", w.writerName, w.batches[id], err)
					}
				}

				// wait for the rate limiter so that not too many batches are inserted at a time
				// ClickHouse recommends to send 1 batch per second, but it appears to work just fine for 5 batches per second
				if err := w.limiter.Wait(w.db.GetContext()); err != nil {
					return fmt.Errorf("rate limiter wait for %q (batch_size %d): %w", w.writerName, w.batches[id], err)
				}

				// send batch
				if err := w.sendBatch(batch); err != nil {
					return fmt.Errorf("sending batch for %q (batch_size %d): %w", w.writerName, w.batches[id], err)
				}

				// if progress updates are enabled, send the number of records
				// this batch handled on the progress channel
				if w.withProgress {
					w.ProgChannel <- batchCount
				}

				// update worker state batch count and alert other workers that this
				// worker is empty
				w.mu.Lock()
				w.batches[id] = 0
				w.cond.Broadcast()
				w.mu.Unlock()
				// reset count and reuse the items buffer's backing array to avoid
				// reallocating a new slice for every batch
				batchCount = 0
				items = items[:0]
			}
		}

		// handle batch when number of items is less than the batch size
		if batchCount > 0 {
			batch, err := conn.PrepareBatch(chCtx, w.query)
			if err != nil {
				return fmt.Errorf("preparing final batch for %q (batch_size %d): %w", w.writerName, w.batches[id], err)
			}

			for _, item := range items {
				if err := batch.AppendStruct(item); err != nil {
					return fmt.Errorf("appending row to final batch for %q (batch_size %d): %w", w.writerName, w.batches[id], err)
				}
			}

			if err := w.sendBatch(batch); err != nil {
				return fmt.Errorf("sending final batch for %q (batch_size %d): %w", w.writerName, w.batches[id], err)
			}

			if w.withProgress {
				w.ProgChannel <- batchCount
			}
		}
		return nil
	})
}
