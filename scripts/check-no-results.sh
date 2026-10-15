#!/bin/sh
# Fails when result numbers are tracked by git. Results stay local until they
# have been reviewed and the compared projects' maintainers have been invited
# to correct their apps; see README.md.
#
#   scripts/check-no-results.sh       checks the repository in the current directory
set -eu

bad="$(git ls-files | grep -E '^results/|-load/|-effort\.json$|summary\.json$' || true)"
stamped="$(git grep -l -F 'local, not publishable' -- ':!docs' ':!README.md' ':!scripts' ':!internal' ':!cmd' || true)"
if [ -n "$bad$stamped" ]; then
	echo "check-no-results: result files must not be committed:" >&2
	printf '%s\n%s\n' "$bad" "$stamped" | sed '/^$/d' | sort -u | sed 's/^/  /' >&2
	exit 1
fi
