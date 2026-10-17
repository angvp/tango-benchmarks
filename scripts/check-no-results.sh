#!/bin/sh
# Fails when raw result files are tracked by git. The numbers worth sharing are
# copied by hand into docs/results.md, with their one-laptop label; the files
# the load generator writes stay local.
#
#   scripts/check-no-results.sh       checks the repository in the current directory
set -eu

bad="$(git ls-files | grep -E '^results/|-load/|-effort\.json$|summary\.json$' || true)"
stamped="$(git grep -l -F 'local, not publishable' -- ':!docs' ':!README.md' ':!scripts' ':!internal' ':!cmd' ':!*_test.go' || true)"
if [ -n "$bad$stamped" ]; then
	echo "check-no-results: result files must not be committed:" >&2
	printf '%s\n%s\n' "$bad" "$stamped" | sed '/^$/d' | sort -u | sed 's/^/  /' >&2
	exit 1
fi
