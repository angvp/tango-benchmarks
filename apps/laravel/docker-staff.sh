#!/bin/sh
# Creates the staff account (the migrations were applied by the image's autorun script).
set -eu
cd /var/www/html
php artisan tinker --execute='\App\Models\User::updateOrCreate(["email" => "bench-staff@example.test"], ["name" => "bench-staff", "password" => \Illuminate\Support\Facades\Hash::make(getenv("BENCH_STAFF_PASSWORD") ?: "bench-staff-password")]);'
