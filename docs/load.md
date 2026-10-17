# Load method

The performance track measures how each pet shop behaves under a fixed, repeatable load. It is never blended with the effort track or the capability matrix.

## The tool

`cmd/load` is the one load generator for every app. It lives in this repository, so its version is the commit you check out; there is nothing to install. It drives a **closed loop**: a fixed number of workers each send a request, wait for the answer, and send the next. Latency is the time from sending the request to reading the whole response. Because the loop is closed, a slow server slows the offered load (coordinated omission): read the percentiles as "what a client sees at this concurrency", not as an open-loop tail.

## Scenarios

All parameters are in `load/params.json`: warm-up, duration, concurrency, number of runs, the seed sizes and the read share. Change the file, not a command line.

| Scenario | What it sends |
|---|---|
| `list` | `GET /api/pets` with a random category, `status=available`, a page from 1 to 5 |
| `pet` | `GET /api/pets/{id}` for a random seeded pet |
| `order` | `POST /api/orders` with a random pet and customer, bearer token from one login before the run |
| `admin` | the admin's pet list HTML, signed in once through its form. N/A, with a reason, for an app without an admin |
| `mix` | `read_percent` reads (half list, half pet) and the rest order POSTs |

Login is checked by the conformance suite, not loaded.

## A run

For each scenario and each of `runs` repetitions (at least 5):

1. reset the database to the seed;
2. restart the app container and time how long until the pet list answers `200` (the cold start);
3. warm up for `warmup_seconds`, then measure for `duration_seconds`, sampling the container's memory every second (the peak is reported).

Only successful responses (`200`, or `201` for orders) count in the latency percentiles and in req/s. Anything else is an error, counted by status and shown in the summary. A framework that rate limits its admin (uAdmin) therefore shows its `429`s instead of a flattering number; its limit is not raised.

The summary shows, per scenario, the median over runs of req/s, p50, p95 and p99, memory and cold start, with the minimum, maximum and the range as a share of the median. A large range means the runs disagree.

## Limits and fairness

Every app and PostgreSQL 17 run with the same container limits (2 CPUs, 2 GiB) from the same `compose.yaml`. Each app uses its framework's documented production server with workers equal to the CPU limit, one recorded configuration, no tuning sweeps.

The load generator runs on the same machine as the containers. That is acceptable for checking the harness and for finding gaps, and the numbers it gives are from one machine, not a controlled run: every result file carries the stamp `local, not publishable`, and [results.md](results.md) says so beside every table. A run on dedicated cloud machines is declared as such.

## Running it

```
./bench.sh load <app>      # about 10 minutes; writes results/<app>-load/ (git-ignored)
```
