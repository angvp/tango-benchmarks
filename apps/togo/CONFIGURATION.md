# togo pet shop: configuration

Built on **togo kernel v0.21.1** with the CLI **v0.2.37** (`togo new`), the scaffold's plugins as installed (auth, auth-dev, cache, data, i18n, queue, realtime, settings, storage, db-postgres). Documented defaults except where this page says otherwise.

| Setting | Value |
|---|---|
| Go | the toolchain in the build image (`golang:1.27-alpine`) |
| Database | PostgreSQL 17 through the `db-postgres` plugin (`DB_DRIVER=pgx`, `DATABASE_URL`) |
| Schema | `internal/db/schema/*.sql`, applied by the scaffold's `cmd/migrate` before the server starts (the container entrypoint runs both) |
| Server | `cmd/api`: the kernel's chi router with Huma (REST and OpenAPI) and gqlgen (GraphQL), on `ADDR` (`:8080`) |
| `APP_ENV` | `production`, which makes the auth plugin require `AUTH_SECRET`; a throwaway benchmark value is set in `compose.yaml` |
| API | the contract's endpoints are hand-written Huma handlers in `internal/rest`, over `database/sql` |
| API auth | `POST /api/login` against a `staff` table with the same opaque HMAC token as the other Go apps; the auth plugin's own routes are installed but not used |
| Generator | `togo make:resource` was run on a scratch project and not used: it generates uuid ids, a plain-array list and problem+json errors, none of which the contract allows |
| Frontend | the scaffold's `web/` directory is kept in the repository and left out of the image (`.dockerignore`); the API is served without it |
| Admin | none for the application's models (see `app.json`) |
| Container limits | 2 CPUs, 2 GiB, the same as every other app |
