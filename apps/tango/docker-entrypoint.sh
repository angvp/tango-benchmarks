#!/bin/sh
# The documented deployment order: apply migrations, make sure the staff
# account exists, then serve. TANGO_ADMIN_PASSWORD is a benchmark credential.
set -eu
/petshop -migrate
/petshop -tango-admin-create=bench-staff || true # already exists on a restart
exec /petshop
