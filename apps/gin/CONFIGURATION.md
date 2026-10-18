# Gin pet shop: configuration

Built on **Gin v1.12.0**, documented defaults except where this page says otherwise. It is the plain Go baseline with Gin in place of chi: the same sqlc queries, the same pgx pool, the same token scheme and the same schema, so the Gin row differs from the baseline row in the router and the framework's own request handling.

| Setting | Value |
|---|---|
| Go | the toolchain in the build image (`golang:1.27-alpine`), static binary in a distroless image |
| Database | PostgreSQL 17 through pgx and sqlc, as in the baseline; Gin has no data layer. The schema is created at start |
| Mode | `GIN_MODE=release`, which Gin's documentation says to use in production |
| Server | `gin.Default()` and `net/http`'s server, so Gin's request logger and panic recovery stay on, as Fuego's default logging does |
| API | handlers take `*gin.Context`; the token check is route-group middleware |
| Validation | the contract's 422s are written by hand; `ShouldBindJSON` only decodes the body |
| Errors | `{"error": ...}` written by a helper |
| API auth | `POST /api/login` against a `staff` table with the same opaque HMAC token as the baseline |
| Admin | none: Gin is an HTTP framework |
| Container limits | 2 CPUs, 2 GiB, the same as every other app |
