package analysis

import (
	"testing"

	"github.com/activecm/rita/v5/config"
)

// BenchmarkCalculateBucketedScore covers the graduated-score interpolation that
// runs for every scored indicator (long connections, C2, etc.) during analysis.
func BenchmarkCalculateBucketedScore(b *testing.B) {
	thresholds := config.ScoreThresholds{
		Base: 1 * 3600,
		Low:  4 * 3600,
		Med:  8 * 3600,
		High: 12 * 3600,
	}
	// spread values across every bucket so the branch predictor sees a realistic mix
	values := []float64{0, 1800, 3600, 5 * 3600, 9 * 3600, 24 * 3600}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateBucketedScore(values[i%len(values)], thresholds)
	}
}
