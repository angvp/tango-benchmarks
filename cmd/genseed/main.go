// Command genseed rewrites seed/data from the generator in internal/seed.
package main

import (
	"fmt"
	"os"

	"github.com/angvp/tango-benchmarks/internal/seed"
)

func main() {
	dir := "seed/data"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := seed.WriteCSV(dir); err != nil {
		fmt.Fprintln(os.Stderr, "genseed:", err)
		os.Exit(1)
	}
}
