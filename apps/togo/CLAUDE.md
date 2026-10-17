# shop — togo project

A togo app (Go API + web, one binary). Built with the togo
microkernel + plugins. This file is loaded by Claude Code — follow it.

## Auth and administration

The auth plugin (`togo-framework/auth`) owns sign-in, sessions, CSRF and user
administration. Admin routes live at `/api/auth/admin/*`; they require the
`admin` role, re-checked against the database on every request, and refuse
API tokens and impersonated sessions. Don't add a second admin API to this app:
extend the plugin or call its routes. With the Next.js frontend, the dashboard
plugin (`togo-framework/dashboard`) provides the admin UI and mail settings. The
TanStack frontend has its own Users page on that API (`web/src/routes/users.tsx`)
and the `/reset-password` page that reset links open; it has no mail settings UI.

Autopilot (`togo-framework/autopilot`) is not installed by default. It is
optional: `togo install togo-framework/autopilot`.

## Conventions (see `.claude/rules/conventions.md`)

- **Add entities via the generator**, never by hand:
  `togo make:resource <Name> <field:type…>` then `togo generate`.
- Edit generated *fragment* bodies (resolvers, handlers); never the `*.gen.go`
  registries (regenerated from `togo.resources.yaml`).
- **REST** = Huma (`internal/rest`), **GraphQL** = gqlgen (`internal/graph`),
  queries = sqlc (`internal/db/queries`), migrations = Atlas (`db/atlas/schema`).
- Keep changes minimal and focused; match the surrounding code.

## Run

```bash
togo generate && togo migrate && togo serve   # http://localhost:8080
```
