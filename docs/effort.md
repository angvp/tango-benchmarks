# Effort rules

The effort track measures what it takes to build the pet shop with a framework. It is reported next to performance and the capability matrix and is never combined with them.

## What is counted

Every file in an app directory is classified once:

- **Infrastructure** (never counted): `Dockerfile`, `docker-entrypoint.sh`, `compose.yaml`, `app.json`, `seed-map.json`, `effort.json`, `steps.txt`, `CONFIGURATION.md`, `.gitignore`, `go.mod`, `go.sum`, `requirements.txt`, plus `__pycache__`, `bin`, `node_modules`, `.git` and `.pyc` files. These exist to run the benchmark, not to build the shop.
- **Generated**: matches a `generated` glob in the app's `effort.json` (tool output such as sqlc code, migrations, scaffold boilerplate nobody edits).
- **Scaffold, edited**: matches a `scaffold_edited` glob (a file the framework's scaffold created and a developer then changed).
- **Hand-written**: everything else.

Each column reports files, non-blank lines and tokens (`cl100k_base`, pinned through `tiktoken-go`, offline). Binary files are skipped.

**Authored** is hand-written plus scaffold-edited. It is shown for convenience; it is not a score and is never ranked.

## Dependencies and image size

Direct dependencies come from `go.mod` or `requirements.txt` (indirect ones are not counted). Image size is the built `tangobench-app` image as Docker reports it.

## Steps

`steps.txt` lists the commands a developer types, from an empty directory to a working pet shop whose staff member can log in. Blank lines and `#` comments do not count. `cd` and `mkdir` are excluded, hand-editing a file is not a command, and environment values are written inline on the command that needs them.

## Running it

```
./bench.sh effort <app>     # writes results/<app>-effort.json (git-ignored)
```

`effort.json` rules with unknown keys or invalid globs are refused. Rules for each app are reviewed in the open; a maintainer who thinks a file is classified unfairly can open an issue and the numbers are redone.
