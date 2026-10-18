# Sinatra pet shop: configuration

Built on **Sinatra 4.2.1** with Ruby 3.4 and Puma, using a modular `Sinatra::Base` app. Sinatra has no data layer, so the app uses Sequel; Rails and Django bring their own.

| Setting | Value |
|---|---|
| Framework | `sinatra 4.2.1`, `rackup 2.3.1` |
| Ruby | 3.4 (`ruby:3.4-slim`) |
| Database layer | `sequel 5.109.0` with `pg 1.7.0`, a pool of 3 connections per worker (the thread count); tables are created at start by a setup script that runs once before the workers start |
| Server | `puma 8.0.2`, 2 worker processes (the CPU limit) of 3 threads each, as for the Rails app |
| Environment | `RACK_ENV=production` and `set :environment, :production` |
| Validation | the contract's 400s and 422s are written by hand; Sinatra reads `params` and has no validation layer |
| API | `get`, `post` and `put` routes; JSON is written with the standard library (the `json` helper is in sinatra-contrib, which is not used) |
| API auth | `POST /api/login` against a `staff` table (`bcrypt 3.1.22`) and the same opaque HMAC token as the other apps; the secret is set once in the entrypoint so all workers accept each other's tokens |
| Admin | none: Sinatra is a routing library |
| Container limits | 2 CPUs, 2 GiB, the same as every other app |
