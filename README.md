# tango-benchmarks

The same small application, a pet shop, built on [tanGO](https://github.com/angvp/tango) and on other web frameworks, to see honestly where each one stands. The aim is understanding, not a leaderboard: there is no "fastest" claim here, and no single score.

Three tracks are kept apart and never blended:

- **Performance:** throughput, latency percentiles, memory and cold start under the same load, limits and database.
- **Effort:** the steps, files, lines and tokens it takes to build the pet shop, and what each framework generates for you.
- **Capability:** what each framework gives you out of the box, with a link to its own documentation for every claim.

Results are published in [`docs/results.md`](docs/results.md), labelled as what they are: one laptop, not a controlled run. The apps have not yet been reviewed by the maintainers of the compared projects, who are welcome to correct them. Only released versions of each framework are measured.

## The pet shop

[`contract/petshop.md`](contract/petshop.md) is the whole specification: four models, a JSON API, and an optional admin. A framework that cannot do a part says so with a reason (recorded as N/A, never as zero). The seed data in [`seed/data`](seed/data) is generated without randomness, so every app starts from the same 20 categories, 10,000 pets and 1,000 customers.

## Running an app

Needs Docker and Go.

```sh
./bench.sh run go-baseline     # build, start PostgreSQL and the app, load the seed, run the suite
./bench.sh down
```

Every app runs in a container with the same CPU and memory limits, against PostgreSQL 17 with the same limits. The conformance suite ([`cmd/conform`](cmd/conform)) checks the contract over HTTP, black box, and must pass before any measurement means anything. It expects a freshly seeded database: `./bench.sh reset <app>` reloads it.

## Methodology

- **Same application, same limits.** Every app and PostgreSQL 17 run with 2 CPUs and 2 GiB. Each framework uses its documented production server on a pinned released version, with its defaults and no tuning sweeps; workers equal the CPU limit.
- **Conformance first.** An app is measured only after it passes the black-box suite on a freshly seeded database.
- **Performance** ([`docs/load.md`](docs/load.md)): five scenarios, parameters in [`load/params.json`](load/params.json), at least five runs each, reported as medians with the run-to-run range, plus peak memory and cold start. Only successful responses count in latency; failures are counted by status.
- **Effort** ([`docs/effort.md`](docs/effort.md)): steps, files, lines and tokens (`cl100k_base`), generated code shown separately, plus dependencies and image size. "Authored" is not a score.
- **Capability** ([`docs/capabilities.md`](docs/capabilities.md)): a fixed list of 22 capabilities, answered yes, partial or no from each framework's own documentation, each with a link or "not documented". No cell is derived from the numbers.
- **Not a leaderboard.** The tracks are never combined, and a framework that lacks a feature gets N/A with a reason, not zero. Runs on a laptop are stamped `local, not publishable` in the result files, and the numbers published here say plainly that they come from one laptop. A run on dedicated cloud machines is still to be made.
- **Corrections welcome.** If you maintain a compared project and think its app is not idiomatic, open an issue or a pull request; the runs are redone after a correction.

Pinned versions: tanGO v0.5.0, Gin v1.12.0, FastAPI 0.143.3 with SQLModel 0.0.48 and uvicorn on Python 3.13, Sinatra 4.2.1 with Sequel 5.109.0 and Puma 8.0.2 on Ruby 3.4, Django 6.1.2 with gunicorn 26.2.0 and psycopg 3.3.6 on Python 3.13, uAdmin v0.10.2, Go with chi, pgx and sqlc 1.28.0 (run through its Docker image), Fuego v0.18.0, togo kernel v0.21.1 with its CLI v0.2.37, Rails 8.1.4 on Ruby 3.4 with Puma, Laravel 13.37 on PHP 8.4 (nginx and PHP-FPM from the `serversideup/php` image), PostgreSQL 17, tokenizer `cl100k_base` via tiktoken-go v0.1.7. Each app's `CONFIGURATION.md` says what was set and why.

## Reproducing a run

```sh
./bench.sh run <app>        # build, start, seed, check the contract
./bench.sh effort <app>     # steps, files, lines, tokens, image size -> results/<app>-effort.json
./bench.sh load <app>       # all load scenarios, about 10 minutes -> results/<app>-load/
./bench.sh down
```

`<app>` is `go-baseline`, `tango`, `django`, `uadmin`, `fuego`, `togo`, `rails`, `laravel`, `gin`, `fastapi` or `sinatra`. Results go to the git-ignored `results/`; a CI check refuses to let them be committed.

## Layout

| Path | What |
|---|---|
| `contract/` | the pet shop contract |
| `seed/data/` | the shared seed, regenerated with `go run ./cmd/genseed` |
| `cmd/conform`, `internal/conform` | the conformance suite |
| `cmd/loadseed`, `internal/loadseed` | loads the seed into an app's own tables, from its `seed-map.json` |
| `internal/fakeapp` | an in-memory pet shop with switchable defects, used to test the suite itself |
| `cmd/effort`, `internal/effort` | the effort counter |
| `cmd/load`, `internal/load` | the load generator and its summary |
| `docs/` | method pages and the capability matrix |
| `apps/<name>/` | one pet shop per framework, with its `app.json`, `seed-map.json` and `compose.yaml` |

An app directory is self-contained: adding a framework means adding `apps/<name>/` and passing the suite.

## Status

The eleven apps (tanGO, Django, uAdmin, a plain Go baseline, Fuego, togo, Rails, Laravel, Gin, FastAPI and Sinatra) pass the conformance suite. Each has its 22 capability answers in [`docs/capabilities.md`](docs/capabilities.md), taken from the framework's own documentation. Results are in [`docs/results.md`](docs/results.md).

Licensed under the MIT license.
