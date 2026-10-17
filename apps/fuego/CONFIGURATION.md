# Fuego pet shop: configuration

Built on **Fuego v0.18.0**, documented defaults except the error serializer (needed for the contract's `{"error": ...}` body).

| Setting | Value |
|---|---|
| Go | the toolchain in the build image (`golang:1.24-alpine`), static binary in a distroless image |
| Database | PostgreSQL 17 through pgx and the same sqlc store as the plain Go baseline; Fuego has no data layer. The schema is created at start |
| Server | `fuego.NewServer(...)` and `Run()`, with Fuego's default logging, OpenAPI generation and Swagger UI left on |
| API | typed handlers (`fuego.Get/Post/Put`); the token check is per-route middleware |
| Validation | the contract's 422s are written by hand; Fuego's own body validation answers 400 |
| Errors | `fuego.WithErrorSerializer` writes `{"error": ...}` |
| API auth | `POST /api/login` against a `staff` table with the same opaque HMAC token as the baseline |
| Admin | none: Fuego is an API framework |
| Container limits | 2 CPUs, 2 GiB, the same as every other app |
