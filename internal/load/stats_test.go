package load

import (
	"testing"
	"time"
)

func TestPercentileUsesNearestRank(t *testing.T) {
	var d []time.Duration
	for i := 1; i <= 100; i++ {
		d = append(d, time.Duration(i)*time.Millisecond)
	}
	for _, c := range []struct {
		p    float64
		want time.Duration
	}{{50, 50 * time.Millisecond}, {95, 95 * time.Millisecond}, {99, 99 * time.Millisecond}, {100, 100 * time.Millisecond}, {0, time.Millisecond}} {
		if got := Percentile(d, c.p); got != c.want {
			t.Errorf("p%v = %v, want %v", c.p, got, c.want)
		}
	}
	if Percentile(nil, 50) != 0 {
		t.Error("no samples must give 0")
	}
}

func TestSpreadOfReportsMedianExtremesAndRange(t *testing.T) {
	s, err := SpreadOf([]float64{100, 120, 80, 110, 90})
	if err != nil {
		t.Fatal(err)
	}
	if s.Median != 100 || s.Min != 80 || s.Max != 120 || s.RangePct != 40 {
		t.Errorf("got %+v", s)
	}
	even, _ := SpreadOf([]float64{10, 20})
	if even.Median != 15 {
		t.Errorf("even median = %v, want 15", even.Median)
	}
	if _, err := SpreadOf(nil); err == nil {
		t.Error("an empty list must be an error")
	}
}

func TestParseMemUsage(t *testing.T) {
	for in, want := range map[string]float64{"123.5MiB / 2GiB": 123.5 * 1048576, "900KiB / 2GiB": 900 * 1024, "1.5GiB / 2GiB": 1.5 * 1073741824, "10B / 1GiB": 10} {
		got, err := ParseMemUsage(in)
		if err != nil || got != want {
			t.Errorf("%q = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "n/a", "12 parsecs / 1GiB"} {
		if _, err := ParseMemUsage(bad); err == nil {
			t.Errorf("%q must be an error", bad)
		}
	}
}
