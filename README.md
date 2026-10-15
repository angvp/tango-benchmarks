# tango-benchmarks

The same small application, a pet shop, built on [tanGO](https://github.com/angvp/tango) and on other web frameworks, to see honestly where each one stands. The aim is understanding, not a leaderboard: there is no "fastest" claim here, and no single score.

Three tracks are kept apart and never blended:

- **Performance:** throughput, latency percentiles, memory and cold start under the same load, limits and database.
- **Effort:** the steps, files, lines and tokens it takes to build the pet shop, and what each framework generates for you.
- **Capability:** what each framework gives you out of the box, with a link to its own documentation for every claim.

Result numbers are not published here until they have been reviewed and the maintainers of the compared projects have been invited to correct their apps. Only released versions of each framework are measured.

## The pet shop

[`contract/petshop.md`](contract/petshop.md) is the whole specification: four models, a JSON API, and an optional admin. A framework that cannot do a part says so with a reason (recorded as N/A, never as zero). The seed data in [`seed/data`](seed/data) is generated without randomness, so every app starts from the same 20 categories, 10,000 pets and 1,000 customers.

## Running an app

Needs Docker and Go.

```sh
./bench.sh run go-baseline     # build, start PostgreSQL and the app, load the seed, run the suite
./bench.sh down
```

Every app runs in a container with the same CPU and memory limits, against PostgreSQL 17 with the same limits. The conformance suite ([`cmd/conform`](cmd/conform)) checks the contract over HTTP, black box, and must pass before any measurement means anything. It expects a freshly seeded database: `./bench.sh reset <app>` reloads it.

## Layout

| Path | What |
|---|---|
| `contract/` | the pet shop contract |
| `seed/data/` | the shared seed, regenerated with `go run ./cmd/genseed` |
| `cmd/conform`, `internal/conform` | the conformance suite |
| `cmd/loadseed`, `internal/loadseed` | loads the seed into an app's own tables, from its `seed-map.json` |
| `internal/fakeapp` | an in-memory pet shop with switchable defects, used to test the suite itself |
| `apps/<name>/` | one pet shop per framework, with its `app.json`, `seed-map.json` and `compose.yaml` |

An app directory is self-contained: adding a framework means adding `apps/<name>/` and passing the suite.

## Status

Phase 1 is under way: tanGO, Django, uAdmin and a plain Go baseline. Fuego, togo, Rails and Laravel follow.

Licensed under the MIT license.
