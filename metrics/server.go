package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server serves RITA's Prometheus metrics and a liveness endpoint.
//
//	/metrics  - Prometheus exposition of the RITA Registry
//	/healthz  - returns 200 OK for container/orchestrator liveness checks
type Server struct {
	httpServer *http.Server
}

// NewServer builds (but does not start) a metrics/health server bound to addr
// (e.g. ":2112").
func NewServer(addr string) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler(),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

// handler builds the metrics + health mux. It is separated out so tests can
// exercise the endpoints without binding a real listener.
func handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(Registry, promhttp.HandlerOpts{Registry: Registry}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

// Start runs the server in a background goroutine and returns a channel that
// receives a non-nil error if the server fails for a reason other than a clean
// shutdown.
func (s *Server) Start() <-chan error {
	errc := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
		}
		close(errc)
	}()
	return errc
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
