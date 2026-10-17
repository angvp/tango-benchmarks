# Results

**One laptop, not a controlled run.** Every number below comes from a single Mac running the load generator, the app and PostgreSQL 17 on the same machine, with each app and the database held to 2 CPUs and 2 GiB. Another machine will give other numbers. Read the shape, not the third digit. A run on dedicated cloud machines has not been made yet, and these tables will be replaced by it when it has.

The apps are ours, written from the [contract](../contract/petshop.md) and not yet reviewed by the maintainers of the compared projects. If you maintain one and think its app is not idiomatic or is badly configured, open an issue or a pull request and the runs are redone. Only released versions are measured, each in its documented production server, with one recorded configuration (see each app's `CONFIGURATION.md`) and no tuning sweeps.

The scenarios are described in [load.md](load.md): `list` is a filtered, paginated pet list, `pet` a single read by id, `order` a `POST` that creates an order, `admin` the server-rendered admin's pet list after one sign-in, and `mix` a blend of reads (list and pet) and order writes. Each scenario was run five times and the table shows the median. The percentage beside a throughput figure is the spread between the slowest and fastest run as a share of the median. A large one, such as 30% or more, means the five runs disagree and the median is soft.

The tracks are not blended and this page ranks nothing: a framework that is faster here may give you less to start with, and the [effort](effort.md) and [capability](capabilities.md) tracks say what. `n/a` means the app has no admin to measure; `none` means the admin was measured and gave no usable result (every request was refused, so there is no throughput to report).

## Throughput

Requests per second, higher is better.

| App | list | pet | order | admin | mix |
|---|---:|---:|---:|---:|---:|
| Go baseline (chi, pgx, sqlc) | 3,822 (40%) | 22,589 (5%) | 10,126 (4%) | n/a | 6,936 (3%) |
| tanGO v0.5.0 | 3,138 (34%) | 20,707 (17%) | 9,746 (36%) | 1,418 (11%) | 6,427 (9%) |
| tanGO v0.4.0 | 611 (12%) | 1,352 (3%) | 747 (4%) | 536 (4%) | 799 (4%) |
| Fuego v0.18.0 | 3,569 (33%) | 18,984 (24%) | 8,871 (18%) | n/a | 6,882 (7%) |
| togo, 32 idle connections | 3,019 (50%) | 23,199 (3%) | 11,014 (2%) | n/a | 6,472 (3%) |
| togo, as scaffolded (2 idle connections) | 627 (22%) | 1,351 (21%) | 722 (4%) | n/a | 856 (5%) |
| Django 6.1.2 | 1,334 (10%) | 1,988 (32%) | 1,682 (2%) | 108 (2%) | 1,805 (1%) |
| Rails 8.1.4 | 1,774 (2%) | 2,405 (3%) | 1,885 (1%) | n/a | 2,197 (12%) |
| Laravel 13.37 | 309 (2%) | 324 (2%) | 240 (2%) | n/a | 302 (3%) |
| uAdmin 0.10.2 | 775 (16%) | 1,352 (6%) | 623 (3%) | none | 786 (17%) |

## p95 latency

Milliseconds, lower is better.

| App | list | pet | order | admin | mix |
|---|---:|---:|---:|---:|---:|
| Go baseline (chi, pgx, sqlc) | 63.0 | 1.9 | 4.0 | n/a | 6.6 |
| tanGO v0.5.0 | 67.9 | 2.8 | 5.3 | 30.4 | 19.6 |
| tanGO v0.4.0 | 104.4 | 85.6 | 94.7 | 96.8 | 97.0 |
| Fuego v0.18.0 | 60.1 | 2.4 | 5.9 | n/a | 10.6 |
| togo, 32 idle connections | 71.6 | 2.4 | 4.3 | n/a | 10.0 |
| togo, as scaffolded (2 idle connections) | 103.6 | 82.9 | 92.9 | n/a | 92.7 |
| Django 6.1.2 | 26.8 | 15.5 | 22.4 | 320.1 | 21.1 |
| Rails 8.1.4 | 20.8 | 15.9 | 19.2 | n/a | 17.8 |
| Laravel 13.37 | 177.5 | 174.4 | 193.7 | n/a | 181.1 |
| uAdmin 0.10.2 | 99.7 | 90.1 | 101.5 | none | 100.2 |

## Memory and cold start

Peak resident memory during the `mix` scenario, and the time from restarting the app container until the pet list answers `200`.

| App | Peak memory, MB | Cold start, ms |
|---|---:|---:|
| Go baseline (chi, pgx, sqlc) | 21 | 108 |
| tanGO v0.5.0 | 22 | 57 |
| tanGO v0.4.0 | 23 | 59 |
| Fuego v0.18.0 | 24 | 107 |
| togo, 32 idle connections | 22 | 213 |
| togo, as scaffolded (2 idle connections) | 22 | 211 |
| Django 6.1.2 | 104 | 1,389 |
| Rails 8.1.4 | 201 | 2,002 |
| Laravel 13.37 | 96 | 2,143 |
| uAdmin 0.10.2 | 37 | 215 |

## Effort

What it took to build the pet shop, from the same contract. Authored lines are what a person wrote or edited; generated lines are produced by a tool such as a code generator or a scaffold. Fewer is not a score. See [effort.md](effort.md).

| App | Authored lines | Authored tokens | Generated lines | Dependencies | Image, MB |
|---|---:|---:|---:|---:|---:|
| Go baseline (chi, pgx, sqlc) | 505 | 4,511 | 277 | 3 | 17 |
| tanGO v0.5.0 | 396 | 3,818 | 165 | 3 | 34 |
| Fuego v0.18.0 | 478 | 4,453 | 277 | 3 | 21 |
| togo | 605 | 6,098 | 5,709 | 25 | 59 |
| Django 6.1.2 | 312 | 2,809 | 83 | 3 | 204 |
| Rails 8.1.4 | 151 | 1,519 | 369 | 4 | 624 |
| Laravel 13.37 | 253 | 2,351 | 2,439 | 3 | 593 |
| uAdmin 0.10.2 | 294 | 2,564 | 0 | 1 | 134 |

## Two things worth knowing

**The two-connection default.** `database/sql` keeps two idle connections by default. Under concurrent load that makes a Go app close and reopen PostgreSQL sessions constantly: tanGO v0.4.0 and togo as scaffolded both land near 1,350 requests per second on `pet` for this reason, and the database container is the busy one. tanGO v0.5.0 changed its default to 32 idle connections, and togo with `DB_MAX_IDLE=32` is the second togo row. Both togo rows are shown because the first is what a new user gets and the second is a one-line setting.

**`list` is the noisy scenario.** It is the slowest of the three shared ones for every Go app, because the contract has every page return a `total` of all matches, and the spread between runs is the widest here. Treat differences of 10 to 20% on `list` as noise.

## Reproducing

```
./bench.sh run <app>        # build, start, conform
./bench.sh load <app>       # about 10 minutes
```

Each app's results land in the git-ignored `results/` directory with the stamp `local, not publishable`; the numbers on this page were copied from there by hand.
