# uAdmin pet shop: configuration

Built on **uAdmin v0.10.2** (the latest release), following its README: `uadmin prepare`, models that embed `uadmin.Model`, `uadmin.Register(...)`, `uadmin.StartServer()`. Documented defaults throughout; nothing tuned.

| Setting | Value |
|---|---|
| Framework | `github.com/uadmin/uadmin v0.10.2` |
| Go | the toolchain in the build image (`golang:1.27-bookworm`); uAdmin's SQLite driver needs cgo even when PostgreSQL is used, so the runtime is `debian:bookworm-slim` |
| Database | PostgreSQL 17, selected with `UADMIN_DB_TYPE`/`HOST`/`PORT`/`NAME`/`USER`/`PASSWORD`; the schema is created by uAdmin's own auto-migration at start |
| Server | `uadmin.StartServer()` on port 8080 |
| Templates and static files | copied into the working directory by `uadmin prepare` at build time; uAdmin reads them from disk at runtime |
| API | the contract's endpoints are ordinary handlers on the same mux (uAdmin's own data API, dAPI, has a different shape and is not used) |
| API auth | `uadmin.Login` for the credentials; the session key it returns is the bearer token |
| Staff account | `bench-staff`, created in `main` on first start. uAdmin also creates its documented default `admin`/`admin` account; it is left as the framework ships it |
| Rate limit | uAdmin limits requests to its admin pages to 3 per second per IP by default. Left as is; the suite waits and retries after a `429` and reports that it did |
| Container limits | 2 CPUs, 2 GiB, the same as every other app |
