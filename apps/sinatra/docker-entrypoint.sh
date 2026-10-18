#!/bin/sh
# Creates the tables and the staff account, then serves with Puma: two worker
# processes (the CPU limit) of three threads each, as for the Rails app.
set -eu
TOKEN_SECRET="$(ruby -rsecurerandom -e 'puts SecureRandom.hex(32)')"
export TOKEN_SECRET
bundle exec ruby setup.rb
exec bundle exec puma -b tcp://0.0.0.0:8080 -w "${WEB_CONCURRENCY:-2}" -t 3:3 config.ru
