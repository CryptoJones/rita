// Package metrics exposes RITA's Prometheus instrumentation.
//
// All collectors are registered on a dedicated Registry (rather than the global
// default) so that importing this package has no side effects beyond what RITA
// explicitly wires up, and so tests can register/scrape in isolation. The
// metrics HTTP endpoint is served by Server (see server.go) and is only started
// when the operator opts in via RITA_METRICS_ADDR.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Registry is RITA's dedicated Prometheus registry.
var Registry = prometheus.NewRegistry()

var factory = promauto.With(Registry)

// Result label values for the *_runs_total counters.
const (
	ResultSuccess = "success"
	ResultError   = "error"
)

var (
	// ImportsTotal counts import runs by result (success/error).
	ImportsTotal = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: "rita", Subsystem: "import", Name: "runs_total",
		Help: "Total number of import runs, labeled by result.",
	}, []string{"result"})

	// ImportDuration observes how long import runs take.
	ImportDuration = factory.NewHistogram(prometheus.HistogramOpts{
		Namespace: "rita", Subsystem: "import", Name: "duration_seconds",
		Help:    "Duration of import runs in seconds.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 12), // 1s .. ~1h
	})

	// RecordsImported counts imported records by log type.
	RecordsImported = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: "rita", Subsystem: "import", Name: "records_total",
		Help: "Total number of records imported, labeled by log type.",
	}, []string{"log_type"})

	// AnalysisRunsTotal counts analysis runs by result (success/error).
	AnalysisRunsTotal = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: "rita", Subsystem: "analysis", Name: "runs_total",
		Help: "Total number of analysis runs, labeled by result.",
	}, []string{"result"})

	// WriteErrorsTotal counts database bulk-write failures.
	WriteErrorsTotal = factory.NewCounter(prometheus.CounterOpts{
		Namespace: "rita", Subsystem: "database", Name: "write_errors_total",
		Help: "Total number of database bulk-write errors.",
	})
)

func init() {
	// Expose Go runtime and process metrics alongside RITA's own.
	Registry.MustRegister(collectors.NewGoCollector())
	Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

// ObserveImport records the outcome and duration of an import run.
func ObserveImport(err error, seconds float64) {
	ImportDuration.Observe(seconds)
	if err != nil {
		ImportsTotal.WithLabelValues(ResultError).Inc()
		return
	}
	ImportsTotal.WithLabelValues(ResultSuccess).Inc()
}

// ObserveAnalysis records the outcome of an analysis run.
func ObserveAnalysis(err error) {
	if err != nil {
		AnalysisRunsTotal.WithLabelValues(ResultError).Inc()
		return
	}
	AnalysisRunsTotal.WithLabelValues(ResultSuccess).Inc()
}
