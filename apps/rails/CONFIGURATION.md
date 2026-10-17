# Rails pet shop: configuration

Built on **Rails 8.1.4** (API mode, `rails new --api`) with Ruby 3.4 and Puma, documented defaults except where this page says otherwise.

| Setting | Value |
|---|---|
| Framework | `rails ~> 8.1.4`, `pg`, `puma`, `bcrypt` (for `has_secure_password`) |
| Database | PostgreSQL 17 through Active Record; the schema is created by `db:prepare` at start (the entrypoint) |
| Server | Puma, `WEB_CONCURRENCY=2` worker processes (the CPU limit, as for Django) and the generated default of 3 threads per worker (`RAILS_MAX_THREADS`), which also sizes the connection pool |
| Environment | `RAILS_ENV=production`; `SECRET_KEY_BASE` is a throwaway benchmark value; logs go to standard output at the default `info` level |
| TLS | `force_ssl` and `assume_ssl` are turned off: the generated production config assumes TLS is terminated by a proxy, and there is none here |
| API | `ActionController::API` controllers; strong parameters; Active Record validations answer `422` |
| API auth | `has_secure_password` for the staff account and `ActiveSupport::MessageVerifier` tokens that expire in 12 hours |
| Staff account | `bench-staff`, created by the entrypoint with `rails runner` |
| Admin | none (see `app.json`) |
| Container limits | 2 CPUs, 2 GiB, the same as every other app |
