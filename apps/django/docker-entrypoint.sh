#!/bin/sh
# Apply migrations, make sure the staff account exists, then serve with
# gunicorn, one worker per CPU in the container's limit. The credentials are
# benchmark values.
set -eu
python manage.py migrate --noinput
DJANGO_SUPERUSER_PASSWORD="${BENCH_STAFF_PASSWORD:-bench-staff-password}" \
	python manage.py createsuperuser --noinput --username bench-staff --email bench@example.test || true # exists on a restart
exec gunicorn petshop.wsgi:application --bind 0.0.0.0:8080 --workers "${WEB_CONCURRENCY:-2}"
