#!/bin/sh
# Creates the tables and the staff account once, then serves with
# `fastapi run`, one worker per CPU in the container's limit.
set -eu
TOKEN_SECRET="$(python -c 'import secrets; print(secrets.token_hex(32))')"
export TOKEN_SECRET
python -m app.setup
exec fastapi run app/main.py --host 0.0.0.0 --port 8080 --workers "${WEB_CONCURRENCY:-2}"
