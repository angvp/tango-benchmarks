package load

import (
	"strings"
	"testing"
)

func TestParseParamsAcceptsTheShippedFile(t *testing.T) {
	p, err := ParseParams([]byte(`{"warmup_seconds":5,"duration_seconds":15,"concurrency":32,"runs":5,
"seed":{"categories":20,"pets":10000,"customers":1000},"list_page_size":20,"read_percent":90,
"scenarios":["list","pet","order","admin","mix"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if p.Concurrency != 32 || p.Runs != 5 || len(p.Scenarios) != 5 {
		t.Errorf("got %+v", p)
	}
}

func TestParseParamsRefusesBadInput(t *testing.T) {
	good := `{"warmup_seconds":1,"duration_seconds":1,"concurrency":1,"runs":5,"seed":{"categories":1,"pets":1,"customers":1},"list_page_size":20,"read_percent":90,"scenarios":["list"]}`
	for name, tc := range map[string]struct{ raw, want string }{
		"unknown key":      {strings.Replace(good, `"runs"`, `"runz"`, 1), "unknown field"},
		"too few runs":     {strings.Replace(good, `"runs":5`, `"runs":2`, 1), "at least 5 runs"},
		"unknown scenario": {strings.Replace(good, `"list"]`, `"nope"]`, 1), "unknown scenario"},
		"zero concurrency": {strings.Replace(good, `"concurrency":1`, `"concurrency":0`, 1), "concurrency"},
		"bad read percent": {strings.Replace(good, `"read_percent":90`, `"read_percent":101`, 1), "read_percent"},
	} {
		_, err := ParseParams([]byte(tc.raw))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want it to mention %q", name, err, tc.want)
		}
	}
}
