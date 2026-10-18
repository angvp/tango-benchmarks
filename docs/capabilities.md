# Capability matrix

What each framework itself provides for the pet shop, answered from its own documentation. The capability list is in `capabilities-list.md` and was fixed before any cell was filled. No cell is derived from the benchmark numbers, and the matrix is never blended with them.

- **yes**: the framework provides it and documents it.
- **partial**: it provides part of it; the note says which part.
- **no**: not provided by the framework. The source is `not documented` when there is nothing to link.

Versions: tanGO v0.5.0, Django 6.1, uAdmin v0.10.2, Fuego v0.18.0, togo kernel v0.21.1 (with the plugin versions its pet shop installs), Rails 8.1 and Laravel 13. The plain Go baseline is the standard library, chi, pgx and sqlc, so its answers are about those libraries. "Source only" means the behaviour is in the code at that tag and not in the project's documentation. A capability that a framework's own organisation ships as an installable plugin counts as provided by that framework, and the note says it is a plugin; this matters for togo, whose kernel is small and whose features are plugins. A third-party package does not count.

## tanGO

| # | Capability | Answer | Note | Source |
|---|---|---|---|---|
| 1 | Models with many-to-one relations | partial | Many-to-one foreign keys only: no many-to-many, no reverse accessors. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/limitations.md) |
| 2 | Migrations generated from models | yes | `tango makemigrations` / `migrate`; renames must be declared. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/migrations.md) |
| 3 | Admin interface generated from models | yes | `admin` app registers models; admin must be installed after the apps that register models. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/admin-registration.md) |
| 4 | Per-model admin permissions | partial | Staff and superuser flags only; permission matrices are not planned. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/limitations.md) |
| 5 | User accounts with hashed passwords | yes | bcrypt-hashed accounts for the admin and the optional `accounts` app. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/accounts.md) |
| 6 | Session (cookie) login | yes | Admin session-cookie login; `auth` primitives for the host's own. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/application-auth.md) |
| 7 | Token (bearer) authentication for APIs | yes | `auth/jwt`, HS256 only, no revocation before expiry. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/jwt-auth.md) |
| 8 | JSON API generated from models | no | No generated API; a view reads and writes with `db.Store`. | not documented |
| 9 | Request validation | partial | `Bind` decodes JSON; bounded strings and foreign-key integrity are checked by `db.Store`; no validation rules layer. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/context-and-binding.md) |
| 10 | CSRF protection | partial | CSRF tokens on admin and `accounts` forms; not a middleware for the host's own routes. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/admin-registration.md) |
| 11 | HTML templating | partial | `Context.HTML` renders a `html/template` you supply; no template language of its own. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/context-and-binding.md) |
| 12 | Pagination helper | partial | `db.Query` Limit/Offset and `Store.Count`; no paginator. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/persistence-crud-and-raw-sql.md) |
| 13 | Query filtering helper | partial | `db.Query` Where/Any/OrderBy; nothing reads the request's query string for you. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/persistence-crud-and-raw-sql.md) |
| 14 | Caching layer | yes | `cache` package with adapters. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/cache.md) |
| 15 | Background jobs or scheduler | partial | Interval-based recurring jobs run in the process; no queue, no retries. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/observability.md) |
| 16 | Internationalisation | yes | `i18n` package. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/i18n.md) |
| 17 | Rate limiting | yes | `ratelimit` token bucket, in-memory and per process. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/rate-limiting.md) |
| 18 | Email sending | yes | `mail` sends plain text to one recipient over SMTP. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/mail.md) |
| 19 | WebSockets | yes | `realtime` rooms over WebSockets, single process. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/realtime-websockets.md) |
| 20 | OpenAPI document generation | yes | The `openapi` package builds an OpenAPI 3.1 document from the route registry; `tango openapi` exports it. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/openapi.md) |
| 21 | Test tooling for the framework's own handlers | yes | `testdb` opens a test database per test. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/sqlite-and-postgresql-setup.md) |
| 22 | Project scaffolding command | yes | `tango newproject` and `newapp`. | [link](https://github.com/angvp/tango/blob/v0.5.0/docs/guides/project-structure.md) |

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

## Fuego

| # | Capability | Answer | Note | Source |
|---|---|---|---|---|
| 1 | Models with many-to-one relations | no | Fuego has no data layer; the pet shop uses sqlc. | [link](https://github.com/go-fuego/fuego/blob/v0.18.0/README.md) |
| 2 | Migrations generated from models | no | No migrations. | not documented |
| 3 | Admin interface generated from models | no | No admin. | not documented |
| 4 | Per-model admin permissions | no | No admin. | not documented |
| 5 | User accounts with hashed passwords | no | Not provided; the `Security` helper signs and checks tokens but accounts are yours to write. | not documented |
| 6 | Session (cookie) login | partial | `Security` can put a signed JWT in a cookie and has login and logout handlers. Source only: the documentation site does not cover it. | [link](https://github.com/go-fuego/fuego/blob/v0.18.0/security.go) |
| 7 | Token (bearer) authentication for APIs | yes | `Security` signs and validates JWTs (ES256), reads the `Authorization` header and has role checks (`AuthWall`). Source only: the documentation site does not cover it. | [link](https://github.com/go-fuego/fuego/blob/v0.18.0/security.go) |
| 8 | JSON API generated from models | no | Routes are hand-written; the OpenAPI document is generated from their types, not from models. | [link](https://github.com/go-fuego/fuego/blob/v0.18.0/documentation/docs/guides/openapi.md) |
| 9 | Request validation | yes | `validate` struct tags checked by go-playground/validator. Failures answer 400. | [link](https://github.com/go-fuego/fuego/blob/v0.18.0/documentation/docs/guides/validation.md) |
| 10 | CSRF protection | no | No CSRF protection. | not documented |
| 11 | HTML templating | yes | Renders `html/template`, Gomponents or Templ. | [link](https://github.com/go-fuego/fuego/blob/v0.18.0/documentation/docs/tutorials/rendering/index.md) |
| 12 | Pagination helper | partial | A `page` query parameter can be declared and documented; there is no paginator. | [link](https://github.com/go-fuego/fuego/blob/v0.18.0/README.md) |
| 13 | Query filtering helper | no | Nothing reads or applies filters. | not documented |
| 14 | Caching layer | no | No caching layer. | not documented |
| 15 | Background jobs or scheduler | no | No jobs or scheduler. | not documented |
| 16 | Internationalisation | no | No internationalisation. | not documented |
| 17 | Rate limiting | no | No rate limiting. | not documented |
| 18 | Email sending | no | No email sending. | not documented |
| 19 | WebSockets | no | No WebSockets. | not documented |
| 20 | OpenAPI document generation | yes | OpenAPI 3 document generated from handler and option types. | [link](https://github.com/go-fuego/fuego/blob/v0.18.0/documentation/docs/guides/openapi.md) |
| 21 | Test tooling for the framework's own handlers | yes | Handlers are testable with `net/http/httptest`, and `MockContext` builds a context without a server. | [link](https://github.com/go-fuego/fuego/blob/v0.18.0/documentation/docs/guides/testing.md) |
| 22 | Project scaffolding command | no | No scaffolding command. | not documented |

## togo

| # | Capability | Answer | Note | Source |
|---|---|---|---|---|
| 1 | Models with many-to-one relations | partial | `make:resource` generates a model struct, an sqlc query file and an Atlas schema from field types; the field types have no relation type, so a foreign key is written by hand. | [link](https://github.com/togo-framework/togo/blob/v0.21.1/docs/generators.md) |
| 2 | Migrations generated from models | yes | `make:resource` writes an Atlas schema for each resource and `togo generate` runs the Atlas diff and migrate. | [link](https://github.com/togo-framework/cli/blob/v0.2.37/README.md) |
| 3 | Admin interface generated from models | partial | The `auth` plugin has an admin API for users and the `dashboard` plugin an admin UI for them; no admin is generated for the application's own models. | [link](https://github.com/togo-framework/auth/blob/v0.10.0/README.md) |
| 4 | Per-model admin permissions | partial | Users carry roles and permissions; there is no per-model admin permission. | [link](https://github.com/togo-framework/auth/blob/v0.10.0/README.md) |
| 5 | User accounts with hashed passwords | yes | The `auth` plugin stores bcrypt-hashed passwords. | [link](https://github.com/togo-framework/auth/blob/v0.10.0/README.md) |
| 6 | Session (cookie) login | yes | The `auth` plugin issues sessions; cookie-authenticated writes need a CSRF token. | [link](https://github.com/togo-framework/auth/blob/v0.10.0/README.md) |
| 7 | Token (bearer) authentication for APIs | yes | The `auth` plugin issues JWTs; bearer requests are accepted. | [link](https://github.com/togo-framework/auth/blob/v0.10.0/README.md) |
| 8 | JSON API generated from models | yes | `make:resource` generates REST (OpenAPI) and GraphQL handlers for each resource. | [link](https://github.com/togo-framework/togo/blob/v0.21.1/docs/generators.md) |
| 9 | Request validation | yes | The `validation` plugin reads `validate` struct tags; generated handlers answer 422 with field errors. | [link](https://github.com/togo-framework/togo/blob/v0.21.1/docs/guide/validation.md) |
| 10 | CSRF protection | partial | The `auth` plugin's routes check a CSRF token for cookie-authenticated writes; it is not documented as a middleware for the application's own routes. | [link](https://github.com/togo-framework/auth/blob/v0.10.0/README.md) |
| 11 | HTML templating | no | API-first; the scaffold's frontend is a separate Next.js or TanStack app. | [link](https://github.com/togo-framework/togo/blob/v0.21.1/README.md) |
| 12 | Pagination helper | partial | The `orm` query builder has `Limit`; the generated list returns a plain array. | [link](https://github.com/togo-framework/togo/blob/v0.21.1/docs/guide/orm.md) |
| 13 | Query filtering helper | partial | The `orm` query builder has `Where` and `Order`; nothing reads the request's query string for you. | [link](https://github.com/togo-framework/togo/blob/v0.21.1/docs/guide/orm.md) |
| 14 | Caching layer | yes | The `cache` plugin; memory, file and database drivers are built in. | [link](https://github.com/togo-framework/cache/blob/v0.3.0/docs/index.md) |
| 15 | Background jobs or scheduler | partial | The `queue` plugin dispatches named jobs to handlers in the process; no retries or schedule are documented. | [link](https://github.com/togo-framework/queue/blob/v0.3.0/docs/index.md) |
| 16 | Internationalisation | yes | The `i18n` plugin. | [link](https://github.com/togo-framework/i18n/blob/v0.2.0/docs/index.md) |
| 17 | Rate limiting | partial | The `auth` plugin limits login and register attempts per IP; there is no general limiter. | [link](https://github.com/togo-framework/auth/blob/v0.10.0/SECURITY.md) |
| 18 | Email sending | yes | The `mail` plugin sends over SMTP, with a log driver and a Resend driver. | [link](https://github.com/togo-framework/togo/blob/v0.21.1/docs/guide/mail-notifications.md) |
| 19 | WebSockets | no | The `realtime` plugin is server-sent events only; it does not provide WebSockets. | [link](https://github.com/togo-framework/realtime/blob/v0.2.2/README.md) |
| 20 | OpenAPI document generation | yes | The REST API is an OpenAPI 3.1 document, exported by `togo generate`. | [link](https://github.com/togo-framework/togo/blob/v0.21.1/README.md) |
| 21 | Test tooling for the framework's own handlers | yes | `togo test` and a feature-test harness that boots the real stack. | [link](https://github.com/togo-framework/togo/blob/v0.21.1/docs/guide/testing.md) |
| 22 | Project scaffolding command | yes | `togo new`. | [link](https://github.com/togo-framework/cli/blob/v0.2.37/README.md) |

## Rails

| # | Capability | Answer | Note | Source |
|---|---|---|---|---|
| 1 | Models with many-to-one relations | yes | Active Record `belongs_to` and `has_many`. | [link](https://guides.rubyonrails.org/association_basics.html) |
| 2 | Migrations generated from models | partial | `bin/rails generate model` and `generate migration` write a migration; it is not derived from later model changes. | [link](https://guides.rubyonrails.org/active_record_migrations.html) |
| 3 | Admin interface generated from models | no | Rails has no admin. | not documented |
| 4 | Per-model admin permissions | no | Rails has no admin. | not documented |
| 5 | User accounts with hashed passwords | yes | `has_secure_password` (bcrypt) and the Rails 8 authentication generator. | [link](https://guides.rubyonrails.org/security.html) |
| 6 | Session (cookie) login | yes | The authentication generator creates session and cookie login. | [link](https://guides.rubyonrails.org/security.html) |
| 7 | Token (bearer) authentication for APIs | partial | `authenticate_with_http_token` reads a bearer token; no token issuing or JWT. | [link](https://guides.rubyonrails.org/security.html) |
| 8 | JSON API generated from models | partial | Scaffold generators write controllers with JSON responses; `rails new --api` makes an API-only app. Nothing is generated from the models alone. | [link](https://guides.rubyonrails.org/api_app.html) |
| 9 | Request validation | yes | Active Record validations. | [link](https://guides.rubyonrails.org/active_record_validations.html) |
| 10 | CSRF protection | yes | `protect_from_forgery`, on by default for forms. | [link](https://guides.rubyonrails.org/security.html) |
| 11 | HTML templating | yes | ERB views and layouts. | [link](https://guides.rubyonrails.org/layouts_and_rendering.html) |
| 12 | Pagination helper | partial | `limit` and `offset`; no paginator. | [link](https://guides.rubyonrails.org/active_record_querying.html) |
| 13 | Query filtering helper | partial | `where` and scopes; nothing reads the request's query string for you. | [link](https://guides.rubyonrails.org/active_record_querying.html) |
| 14 | Caching layer | yes | Fragment, low-level and HTTP caching with pluggable stores. | [link](https://guides.rubyonrails.org/caching_with_rails.html) |
| 15 | Background jobs or scheduler | yes | Active Job with retries; recurring jobs come from the default Solid Queue backend. | [link](https://guides.rubyonrails.org/active_job_basics.html) |
| 16 | Internationalisation | yes | The I18n API. | [link](https://guides.rubyonrails.org/i18n.html) |
| 17 | Rate limiting | yes | `rate_limit` in controllers. | [link](https://guides.rubyonrails.org/security.html) |
| 18 | Email sending | yes | Action Mailer. | [link](https://guides.rubyonrails.org/action_mailer_basics.html) |
| 19 | WebSockets | yes | Action Cable. | [link](https://guides.rubyonrails.org/action_cable_overview.html) |
| 20 | OpenAPI document generation | no | No OpenAPI generation. | not documented |
| 21 | Test tooling for the framework's own handlers | yes | Minitest integration, controller and system tests. | [link](https://guides.rubyonrails.org/testing.html) |
| 22 | Project scaffolding command | yes | `rails new` and the generators. | [link](https://guides.rubyonrails.org/command_line.html) |

## Laravel

| # | Capability | Answer | Note | Source |
|---|---|---|---|---|
| 1 | Models with many-to-one relations | yes | Eloquent `belongsTo` and `hasMany`. | [link](https://laravel.com/docs/13.x/eloquent-relationships) |
| 2 | Migrations generated from models | partial | `make:migration` and `make:model -m` write a migration skeleton; the schema is written in it by hand. | [link](https://laravel.com/docs/13.x/migrations) |
| 3 | Admin interface generated from models | no | Laravel has no admin. | not documented |
| 4 | Per-model admin permissions | no | Laravel has no admin. | not documented |
| 5 | User accounts with hashed passwords | yes | The hasher uses bcrypt or Argon2, and the starter kits add registration. | [link](https://laravel.com/docs/13.x/authentication) |
| 6 | Session (cookie) login | yes | Session guard and cookie login. | [link](https://laravel.com/docs/13.x/authentication) |
| 7 | Token (bearer) authentication for APIs | yes | Sanctum API tokens, a first-party package installed with `install:api`. | [link](https://laravel.com/docs/13.x/sanctum) |
| 8 | JSON API generated from models | partial | `make:model --api` and `make:resource` write API resource controllers and JSON resources; the resource fields are written by hand. | [link](https://laravel.com/docs/13.x/eloquent-resources) |
| 9 | Request validation | yes | Validation rules on requests. | [link](https://laravel.com/docs/13.x/validation) |
| 10 | CSRF protection | yes | Request-forgery middleware, on for the web routes. | [link](https://laravel.com/docs/13.x/csrf) |
| 11 | HTML templating | yes | Blade templates. | [link](https://laravel.com/docs/13.x/blade) |
| 12 | Pagination helper | yes | The paginator. | [link](https://laravel.com/docs/13.x/pagination) |
| 13 | Query filtering helper | partial | Query builder `where` and scopes; nothing reads the request's query string for you. | [link](https://laravel.com/docs/13.x/queries) |
| 14 | Caching layer | yes | Cache with several stores. | [link](https://laravel.com/docs/13.x/cache) |
| 15 | Background jobs or scheduler | yes | Queues with retries, and a task scheduler. | [link](https://laravel.com/docs/13.x/queues) |
| 16 | Internationalisation | yes | Localization. | [link](https://laravel.com/docs/13.x/localization) |
| 17 | Rate limiting | yes | `RateLimiter` and the `throttle` middleware. | [link](https://laravel.com/docs/13.x/rate-limiting) |
| 18 | Email sending | yes | Mail over SMTP and API drivers. | [link](https://laravel.com/docs/13.x/mail) |
| 19 | WebSockets | yes | Broadcasting, with Reverb as a first-party WebSocket server. | [link](https://laravel.com/docs/13.x/reverb) |
| 20 | OpenAPI document generation | no | No OpenAPI generation. | not documented |
| 21 | Test tooling for the framework's own handlers | yes | PHPUnit and Pest testing, HTTP tests. | [link](https://laravel.com/docs/13.x/testing) |
| 22 | Project scaffolding command | yes | The `laravel new` installer and `artisan make:` commands. | [link](https://laravel.com/docs/13.x/installation) |
