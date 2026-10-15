package load

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var memUnits = map[string]float64{"B": 1, "kB": 1e3, "KiB": 1 << 10, "MB": 1e6, "MiB": 1 << 20, "GB": 1e9, "GiB": 1 << 30}

// ParseMemUsage reads the used part of `docker stats` output such as
// "123.4MiB / 2GiB" and returns bytes.
func ParseMemUsage(s string) (float64, error) {
	used, _, _ := strings.Cut(s, "/")
	used = strings.TrimSpace(used)
	i := strings.IndexFunc(used, func(r rune) bool { return (r < '0' || r > '9') && r != '.' })
	if i <= 0 {
		return 0, fmt.Errorf("cannot read memory from %q", s)
	}
	n, err := strconv.ParseFloat(used[:i], 64)
	unit, ok := memUnits[used[i:]]
	if err != nil || !ok {
		return 0, fmt.Errorf("cannot read memory from %q", s)
	}
	return n * unit, nil
}

// PeakMemory samples a container's memory once a second until ctx ends and
// returns the highest reading in bytes.
func PeakMemory(ctx context.Context, container string) (float64, error) {
	var peak float64
	var lastErr error
	for ctx.Err() == nil {
		out, err := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", "{{.MemUsage}}", container).Output()
		if err == nil {
			var b float64
			if b, err = ParseMemUsage(string(out)); err == nil {
				peak = max(peak, b)
			}
		}
		if err != nil && ctx.Err() == nil {
			lastErr = err
		}
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
		}
	}
	if peak == 0 && lastErr != nil {
		return 0, lastErr
	}
	return peak, nil
}

// ColdStart restarts a container and returns how long until the pet list
// answers 200.
func ColdStart(container, readyURL string, limit time.Duration) (time.Duration, error) {
	if out, err := exec.Command("docker", "restart", "-t", "5", container).CombinedOutput(); err != nil {
		return 0, fmt.Errorf("docker restart %s: %w: %s", container, err, out)
	}
	// A short timeout and no keep-alive: right after a restart the published
	// port can accept a connection and then stall, which must not hang the wait.
	probe := &http.Client{Timeout: 2 * time.Second, Transport: &http.Transport{DisableKeepAlives: true}}
	begin := time.Now()
	for time.Since(begin) < limit {
		resp, err := probe.Get(readyURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return time.Since(begin), nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return 0, fmt.Errorf("%s was not ready within %v of a restart", readyURL, limit)
}
