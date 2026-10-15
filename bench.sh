#!/bin/sh
# Runs one pet shop in containers and checks it against the contract.
#
#   ./bench.sh run <app>        up + reset + conform
#   ./bench.sh up <app>         build and start the app and PostgreSQL
#   ./bench.sh reset <app>      truncate and reload the seed (a fresh database)
#   ./bench.sh conform <app>    run the conformance suite, write results/<app>.json
#   ./bench.sh effort <app>     build the image, count steps/files/lines/tokens, write results/<app>-effort.json
#   ./bench.sh down             stop everything
#
# <app> is a directory under apps/. Results go to results/, which git ignores.
set -eu
cd "$(dirname "$0")"

DSN="postgres://bench:bench@127.0.0.1:55440/bench?sslmode=disable"
BASE_URL="http://127.0.0.1:8080"

usage() { sed -n '2,11p' "$0" | sed 's/^# \{0,1\}//'; exit 2; }

compose() {
	app="$1"; shift
	docker compose -f compose.yaml -f "apps/$app/compose.yaml" "$@"
}

need_app() {
	[ -n "${1:-}" ] && [ -f "apps/$1/app.json" ] || { echo "bench.sh: apps/<app>/app.json not found (app: ${1:-none})" >&2; exit 2; }
}

build_effort() { mkdir -p bin && go build -o bin/effort ./cmd/effort; }

build_tools() { mkdir -p bin && go build -o bin/conform ./cmd/conform && go build -o bin/loadseed ./cmd/loadseed; }

wait_ready() {
	i=0
	until curl -fs -o /dev/null "$BASE_URL/api/pets?page_size=1"; do
		i=$((i + 1))
		[ "$i" -lt 60 ] || { echo "bench.sh: the app did not become ready in 60s" >&2; exit 1; }
		sleep 1
	done
}

cmd="${1:-}"; app="${2:-}"
case "$cmd" in
	up)      need_app "$app"; compose "$app" up -d --build --wait db app; wait_ready ;;
	reset)   need_app "$app"; build_tools; ./bin/loadseed -dsn "$DSN" -map "apps/$app/seed-map.json" ;;
	conform) need_app "$app"; build_tools; ./bin/conform -base-url "$BASE_URL" -app "apps/$app/app.json" -out "results/$app.json" ;;
	effort)  need_app "$app"; build_effort; compose "$app" build app; ./bin/effort -app "apps/$app" -image tangobench-app -out "results/$app-effort.json" ;;
	run)     "$0" up "$app"; "$0" reset "$app"; "$0" conform "$app" ;;
	down)    for f in apps/*/compose.yaml; do docker compose -f compose.yaml -f "$f" down -v --remove-orphans >/dev/null 2>&1 || true; done ;;
	*)       usage ;;
esac
