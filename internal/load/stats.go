package load

import (
	"fmt"
	"math"
	"slices"
	"time"
)

// Percentile returns the p-th percentile (0-100) of sorted durations by the
// nearest-rank method. It is 0 for no samples.
func Percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	rank := int(math.Ceil(p / 100 * float64(len(sorted))))
	return sorted[min(max(rank, 1), len(sorted))-1]
}

// Spread describes one metric over several runs: the median, the extremes and
// how far apart they are, as a share of the median.
type Spread struct {
	Median float64 `json:"median"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	// RangePct is (max-min)/median*100; a large value means the runs disagree.
	RangePct float64 `json:"range_pct"`
}

// SpreadOf summarises values; it fails for an empty list.
func SpreadOf(values []float64) (Spread, error) {
	if len(values) == 0 {
		return Spread{}, fmt.Errorf("no values to summarise")
	}
	sorted := slices.Sorted(slices.Values(values))
	n := len(sorted)
	median := sorted[n/2]
	if n%2 == 0 {
		median = (sorted[n/2-1] + sorted[n/2]) / 2
	}
	s := Spread{Median: median, Min: sorted[0], Max: sorted[n-1]}
	if median > 0 {
		s.RangePct = (s.Max - s.Min) / median * 100
	}
	return s, nil
}
