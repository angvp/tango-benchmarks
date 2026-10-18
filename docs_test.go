package tangobenchmarks_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The capability matrix must answer every capability on the fixed list for
// every framework, and every answer must say where it comes from.
func TestCapabilityMatrixIsComplete(t *testing.T) {
	list, err := os.ReadFile("docs/capabilities-list.md")
	if err != nil {
		t.Fatal(err)
	}
	want := regexp.MustCompile(`(?m)^\d+\. (.+)$`).FindAllStringSubmatch(string(list), -1)
	matrix, err := os.ReadFile("docs/capabilities.md")
	if err != nil {
		t.Fatal(err)
	}
	sections := strings.Split(string(matrix), "\n## ")[1:]
	if len(sections) != 11 {
		t.Fatalf("got %d frameworks, want 11", len(sections))
	}
	row := regexp.MustCompile(`^\| (\d+) \| (.+?) \| (yes|partial|no) \| (.+?) \| (not documented|\[link\]\(https://[^)]+\)) \|$`)
	for _, s := range sections {
		name, body, _ := strings.Cut(s, "\n")
		var got []string
		for _, line := range strings.Split(body, "\n") {
			if !strings.HasPrefix(line, "| ") || strings.HasPrefix(line, "| #") {
				continue
			}
			m := row.FindStringSubmatch(line)
			if m == nil {
				t.Errorf("%s: malformed row %q (needs answer yes/partial/no and a link or \"not documented\")", name, line)
				continue
			}
			got = append(got, m[2])
		}
		if len(got) != len(want) {
			t.Errorf("%s: %d capabilities answered, want %d", name, len(got), len(want))
			continue
		}
		for i := range want {
			if got[i] != want[i][1] {
				t.Errorf("%s: row %d is %q, the list says %q", name, i+1, got[i], want[i][1])
			}
		}
	}
}
