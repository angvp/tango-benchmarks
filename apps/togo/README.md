# Shop

Built with [togo](https://github.com/togo-framework) — Go, the artisan way.
API-first (GraphQL + REST/OpenAPI), Supabase auth, plugin-extensible, AI-native.

## Quickstart

```bash
cp .env.example .env          # configure DATABASE_URL + Supabase
docker compose up -d          # Postgres / Supabase
togo make:resource Post title:string body:text:nullable
togo generate                 # sqlc + gqlgen + atlas + openapi
togo migrate
togo serve                    # backend + frontend together
```

`togo serve` runs **both** the Go API and the Next.js frontend (installing web
deps on first run). Use `togo serve --api-only` / `--web-only`, or `togo web`.

- API → http://localhost:8080  (GraphQL `/graphql` · REST `/api` · docs `/docs`)
- Web → http://localhost:3000

## Stack

- **API**: chi + Huma (REST/OpenAPI 3.1) + gqlgen (GraphQL)
- **Data**: sqlc (typed queries) + Atlas (migrations) + pgx/Postgres
- **Auth + Storage**: Supabase (`@supabase/ssr`) — client/server/middleware wired
- **Frontend**: Next.js (App Router) or TanStack + Tailwind CSS v4 + the Nasaq UI kit
  ([`@fadymondy/nasaq`](https://www.npmjs.com/package/@fadymondy/nasaq), ToGO brand), lucide icons
- **AI**: `.claude/` skills + agents + `.mcp.json` wired to the togo MCP server

## Auth and admin

Users, sessions and user administration come from the auth plugin
([`togo-framework/auth`](https://github.com/togo-framework/auth)). The admin API is
`/api/auth/admin/*`: every route needs the `admin` role, checked against the
database on each request. Cookie-authenticated writes send `X-CSRF-Token` from
`GET /api/auth/csrf`. With the Next.js frontend, the dashboard plugin adds the admin
UI (users, impersonation, mail settings). The TanStack frontend ships a Users page
(roles, reset and magic links, impersonation) and the `/reset-password` page that
reset links open (`AUTH_RESET_PATH`).

Accounts start with no roles, and auth has no endpoint that grants the first
admin: set `roles` to `admin` for that user in the `users` table (a seeder or SQL),
or call `SetRoles` from trusted Go code. The `auth-dev` developer login is not an
admin under auth v0.10.0, because roles come from the database, not the token.

Impersonation hands the admin a bearer token for the target user; the TanStack
frontend sends it on API calls instead of the admin's cookie. `/events`
(Server-Sent Events via `EventSource`) cannot carry that header, so the stream keeps
the browser's own session while impersonating. Don't use it to deliver
per-user data.

Optional, not installed by default: Autopilot (`togo install togo-framework/autopilot`).

## Layout

```
cmd/api/            HTTP entrypoint (REST + GraphQL)
internal/models/    domain structs            (togo make:resource)
internal/db/        sqlc schema + queries + generated code
internal/graph/     GraphQL schema + resolvers
internal/rest/      Huma handlers + generated route registry
internal/resources/ resource descriptors (dashboard/admin)
db/atlas/           Atlas desired-state schema + migrations
web/                Next.js frontend
.claude/            skills, agents, rules for Claude Code
togo.yaml           project config
togo.resources.yaml resource manifest (source of truth for codegen)
```
