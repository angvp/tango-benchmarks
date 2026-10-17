#!/bin/sh
# Applies the migrations, creates the staff account, then serves with Puma.
set -eu
bin/rails db:prepare
bin/rails runner 'StaffMember.find_or_initialize_by(username: "bench-staff").update!(password: ENV.fetch("BENCH_STAFF_PASSWORD", "bench-staff-password"))'
exec bundle exec puma -C config/puma.rb
