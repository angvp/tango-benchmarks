#!/bin/sh
# Applies the schema files, then serves.
set -eu
/app/migrate
exec /app/api
