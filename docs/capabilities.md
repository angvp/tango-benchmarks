# Capability matrix

What each framework itself provides for the pet shop, answered from its own documentation. The capability list is in `capabilities-list.md` and was fixed before any cell was filled. No cell is derived from the benchmark numbers, and the matrix is never blended with them.

- **yes**: the framework provides it and documents it.
- **partial**: it provides part of it; the note says which part.
- **no**: not provided by the framework. The source is `not documented` when there is nothing to link.

Versions: tanGO v0.4.0, Django 6.1, uAdmin v0.10.2. The plain Go baseline is the standard library, chi, pgx and sqlc, so its answers are about those libraries. "Source only" means the behaviour is in the code at that tag and not in the project's documentation.

## tanGO

| # | Capability | Answer | Note | Source |
|---|---|---|---|---|
| 1 | Models with many-to-one relations | partial | Many-to-one foreign keys only: no many-to-many, no reverse accessors. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/limitations.md) |
| 2 | Migrations generated from models | yes | `tango makemigrations` / `migrate`; renames must be declared. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/migrations.md) |
| 3 | Admin interface generated from models | yes | `admin` app registers models; admin must be installed after the apps that register models. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/admin-registration.md) |
| 4 | Per-model admin permissions | partial | Staff and superuser flags only; permission matrices are not planned. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/limitations.md) |
| 5 | User accounts with hashed passwords | yes | bcrypt-hashed accounts for the admin and the optional `accounts` app. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/accounts.md) |
| 6 | Session (cookie) login | yes | Admin session-cookie login; `auth` primitives for the host's own. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/application-auth.md) |
| 7 | Token (bearer) authentication for APIs | yes | `auth/jwt`, HS256 only, no revocation before expiry. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/jwt-auth.md) |
| 8 | JSON API generated from models | no | No generated API; a view reads and writes with `db.Store`. | not documented |
| 9 | Request validation | partial | `Bind` decodes JSON; bounded strings and foreign-key integrity are checked by `db.Store`; no validation rules layer. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/context-and-binding.md) |
| 10 | CSRF protection | partial | CSRF tokens on admin and `accounts` forms; not a middleware for the host's own routes. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/admin-registration.md) |
| 11 | HTML templating | partial | `Context.HTML` renders a `html/template` you supply; no template language of its own. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/context-and-binding.md) |
| 12 | Pagination helper | partial | `db.Query` Limit/Offset and `Store.Count`; no paginator. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/persistence-crud-and-raw-sql.md) |
| 13 | Query filtering helper | partial | `db.Query` Where/Any/OrderBy; nothing reads the request's query string for you. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/persistence-crud-and-raw-sql.md) |
| 14 | Caching layer | yes | `cache` package with adapters. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/cache.md) |
| 15 | Background jobs or scheduler | partial | Interval-based recurring jobs run in the process; no queue, no retries. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/observability.md) |
| 16 | Internationalisation | yes | `i18n` package. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/i18n.md) |
| 17 | Rate limiting | yes | `ratelimit` token bucket, in-memory and per process. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/rate-limiting.md) |
| 18 | Email sending | yes | `mail` sends plain text to one recipient over SMTP. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/mail.md) |
| 19 | WebSockets | yes | `realtime` rooms over WebSockets, single process. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/realtime-websockets.md) |
| 20 | OpenAPI document generation | no | No OpenAPI support. | not documented |
| 21 | Test tooling for the framework's own handlers | yes | `testdb` opens a test database per test. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/sqlite-and-postgresql-setup.md) |
| 22 | Project scaffolding command | yes | `tango newproject` and `newapp`. | [link](https://github.com/angvp/tango/blob/v0.4.0/docs/guides/project-structure.md) |

## Django

| # | Capability | Answer | Note | Source |
|---|---|---|---|---|
| 1 | Models with many-to-one relations | yes | ForeignKey, ManyToManyField, OneToOneField. | [link](https://docs.djangoproject.com/en/6.1/topics/db/models/#relationships) |
| 2 | Migrations generated from models | yes | `makemigrations` / `migrate`. | [link](https://docs.djangoproject.com/en/6.1/topics/migrations/) |
| 3 | Admin interface generated from models | yes | `django.contrib.admin`. | [link](https://docs.djangoproject.com/en/6.1/ref/contrib/admin/) |
| 4 | Per-model admin permissions | yes | Add, change, delete and view permissions per model. | [link](https://docs.djangoproject.com/en/6.1/topics/auth/default/#permissions-and-authorization) |
| 5 | User accounts with hashed passwords | yes | `django.contrib.auth`, PBKDF2 by default. | [link](https://docs.djangoproject.com/en/6.1/topics/auth/) |
| 6 | Session (cookie) login | yes | Sessions framework. | [link](https://docs.djangoproject.com/en/6.1/topics/http/sessions/) |
| 7 | Token (bearer) authentication for APIs | no | Not in core; the usual answer is a third-party package. | not documented |
| 8 | JSON API generated from models | no | Not in core; the usual answer is a third-party package. | not documented |
| 9 | Request validation | yes | Forms and model validation. | [link](https://docs.djangoproject.com/en/6.1/ref/forms/validation/) |
| 10 | CSRF protection | yes | `CsrfViewMiddleware`. | [link](https://docs.djangoproject.com/en/6.1/ref/csrf/) |
| 11 | HTML templating | yes | Django template language and Jinja2 backend. | [link](https://docs.djangoproject.com/en/6.1/topics/templates/) |
| 12 | Pagination helper | yes | `Paginator`. | [link](https://docs.djangoproject.com/en/6.1/topics/pagination/) |
| 13 | Query filtering helper | partial | QuerySet `filter()`; nothing maps a request's query string to filters. | [link](https://docs.djangoproject.com/en/6.1/topics/db/queries/#retrieving-specific-objects-with-filters) |
| 14 | Caching layer | yes | Cache framework with several backends. | [link](https://docs.djangoproject.com/en/6.1/topics/cache/) |
| 15 | Background jobs or scheduler | partial | Tasks framework defines and enqueues work; core has no worker and its backends are for development and testing. | [link](https://docs.djangoproject.com/en/6.1/topics/tasks/) |
| 16 | Internationalisation | yes | Translation framework. | [link](https://docs.djangoproject.com/en/6.1/topics/i18n/) |
| 17 | Rate limiting | no | Not in core. | not documented |
| 18 | Email sending | yes | `django.core.mail`. | [link](https://docs.djangoproject.com/en/6.1/topics/email/) |
| 19 | WebSockets | no | ASGI is supported; WebSocket handling is a separate project. | not documented |
| 20 | OpenAPI document generation | no | Not in core. | not documented |
| 21 | Test tooling for the framework's own handlers | yes | Test client and test runner. | [link](https://docs.djangoproject.com/en/6.1/topics/testing/) |
| 22 | Project scaffolding command | yes | `django-admin startproject` and `startapp`. | [link](https://docs.djangoproject.com/en/6.1/ref/django-admin/#startproject) |

## uAdmin

| # | Capability | Answer | Note | Source |
|---|---|---|---|---|
| 1 | Models with many-to-one relations | yes | Foreign keys and many-to-many, listed in the README features. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/README.md) |
| 2 | Migrations generated from models | yes | "Database schema migration" in the README features; it runs when the app starts. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/README.md) |
| 3 | Admin interface generated from models | yes | The admin is the framework's core. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/README.md) |
| 4 | Per-model admin permissions | yes | User, group and per-model permissions. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/README.md) |
| 5 | User accounts with hashed passwords | yes | Hash with salt, password reset and two-factor authentication. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/README.md) |
| 6 | Session (cookie) login | yes | Sessions and a login handler. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/README.md) |
| 7 | Token (bearer) authentication for APIs | partial | The dAPI login returns a session key that later calls send; not a standard bearer token. Source only. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/d_api_login.go) |
| 8 | JSON API generated from models | yes | The Data Access API (dAPI) lists, reads, adds, edits and deletes any model. The benchmark app does not use it because its JSON shapes differ from the contract. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/README.md) |
| 9 | Request validation | yes | "Validation for user input" in the README features. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/README.md) |
| 10 | CSRF protection | yes | A CSRF check on forms. Source only. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/check_csrf.go) |
| 11 | HTML templating | partial | Admin templates and render helpers; no general application templating guide. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/render_html.go) |
| 12 | Pagination helper | yes | Admin lists page; dAPI reads accept limit and offset. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/d_api_read.go) |
| 13 | Query filtering helper | yes | dAPI reads accept filters. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/d_api_read.go) |
| 14 | Caching layer | no | Only an internal permission cache. | not documented |
| 15 | Background jobs or scheduler | no | Only an internal A/B test sync loop. | not documented |
| 16 | Internationalisation | yes | "Multilingual translation" in the README features. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/README.md) |
| 17 | Rate limiting | partial | A per-IP quota on pages and login (3 requests a second by default); not a general limiter. Source only. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/rate_limit.go) |
| 18 | Email sending | yes | Email configuration and sending. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/README.md) |
| 19 | WebSockets | no | No WebSocket support. | not documented |
| 20 | OpenAPI document generation | yes | OpenAPI generation. Source only. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/openapi.go) |
| 21 | Test tooling for the framework's own handlers | no | No test helpers for host handlers. | not documented |
| 22 | Project scaffolding command | partial | The `uadmin` command prepares templates and static files; it does not generate a project. | [link](https://github.com/uadmin/uadmin/blob/v0.10.2/cmd/uadmin/main.go) |

## Plain Go baseline

| # | Capability | Answer | Note | Source |
|---|---|---|---|---|
| 1 | Models with many-to-one relations | partial | sqlc generates typed Go from SQL; there is no model layer. Foreign keys live in the schema. | [link](https://docs.sqlc.dev/en/latest/) |
| 2 | Migrations generated from models | no | sqlc reads migrations written by hand; it does not generate them. | [link](https://docs.sqlc.dev/en/latest/howto/ddl.html) |
| 3 | Admin interface generated from models | no | Not a goal of the libraries. | not documented |
| 4 | Per-model admin permissions | no | Not a goal of the libraries. | not documented |
| 5 | User accounts with hashed passwords | no | `golang.org/x/crypto/bcrypt` hashes a password; accounts are yours to write. | [link](https://pkg.go.dev/golang.org/x/crypto/bcrypt) |
| 6 | Session (cookie) login | no | Not in the libraries. | not documented |
| 7 | Token (bearer) authentication for APIs | no | Not in the libraries. | not documented |
| 8 | JSON API generated from models | no | Not in the libraries. | not documented |
| 9 | Request validation | no | Not in the libraries. | not documented |
| 10 | CSRF protection | yes | `http.CrossOriginProtection` (Go 1.25). | [link](https://pkg.go.dev/net/http#CrossOriginProtection) |
| 11 | HTML templating | yes | `html/template`. | [link](https://pkg.go.dev/html/template) |
| 12 | Pagination helper | no | Write the query. | [link](https://docs.sqlc.dev/en/latest/) |
| 13 | Query filtering helper | no | Write the query. | [link](https://docs.sqlc.dev/en/latest/) |
| 14 | Caching layer | no | Not in the libraries. | not documented |
| 15 | Background jobs or scheduler | no | Not in the libraries. | not documented |
| 16 | Internationalisation | no | Not in the libraries. | not documented |
| 17 | Rate limiting | no | Not in the libraries. | not documented |
| 18 | Email sending | partial | `net/smtp` is frozen and low level. | [link](https://pkg.go.dev/net/smtp) |
| 19 | WebSockets | no | Not in the libraries. | not documented |
| 20 | OpenAPI document generation | no | Not in the libraries. | not documented |
| 21 | Test tooling for the framework's own handlers | yes | `net/http/httptest`. | [link](https://pkg.go.dev/net/http/httptest) |
| 22 | Project scaffolding command | no | `go mod init` and an editor. | not documented |
