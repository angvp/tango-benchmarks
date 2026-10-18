# FastAPI pet shop: configuration

Built on **FastAPI 0.143.3** (`fastapi[standard]`) with Python 3.13, following the documentation's own choices where it has them.

| Setting | Value |
|---|---|
| Framework | `fastapi[standard]==0.143.3` |
| Python | 3.13 (`python:3.13-slim`) |
| Database layer | SQLModel 0.0.48, which the FastAPI documentation's "SQL (Relational) Databases" tutorial uses, over SQLAlchemy with `psycopg[binary]==3.3.6`; synchronous sessions through a dependency, with SQLAlchemy's default pool |
| Tables | created at start with `SQLModel.metadata.create_all`, in a setup step that runs once before the workers start; the foreign keys, the two checks and the `category_id` and `status` indexes match the other apps |
| Server | `fastapi run --workers 2` (uvicorn, the CPU limit, as for Django and Rails); the endpoints are plain `def` handlers, which FastAPI runs in a thread pool |
| Validation | the contract's 422s are written by hand: the body is read as a plain dict so that the 401 and the 404 come before them. FastAPI's own body validation is unused |
| Errors | exception handlers write `{"error": ...}`; a request FastAPI itself rejects (a bad `page`, a malformed body) answers 400, as the contract says, where FastAPI's default is 422 |
| API auth | `POST /api/login` against a `staff` table (pwdlib with Argon2, as in the security tutorial) and the same opaque HMAC token as the other apps, read through `HTTPBearer`; the secret is set once in the entrypoint so both workers accept each other's tokens |
| Documentation routes | the interactive docs and OpenAPI document stay on, as the defaults |
| Admin | none: FastAPI is an API framework |
| Container limits | 2 CPUs, 2 GiB, the same as every other app |
