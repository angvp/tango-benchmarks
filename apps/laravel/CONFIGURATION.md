# Laravel pet shop: configuration

Built on **Laravel 13.37** (`composer create-project`) with Sanctum (`php artisan install:api`) and PHP 8.4, served by nginx and PHP-FPM from the `serversideup/php:8.4-fpm-nginx` image, which is a maintained production image for Laravel. Documented defaults except where this page says otherwise.

| Setting | Value |
|---|---|
| Framework | `laravel/framework ^13.17`, `laravel/sanctum ^4.0`, `laravel/tinker` |
| Database | PostgreSQL 17 through Eloquent and the query builder; the schema is created by `php artisan migrate --force`, run by the image's `AUTORUN_ENABLED` script at start |
| Server | nginx and PHP-FPM, the image's defaults: dynamic process manager, at most 20 children |
| OPcache | enabled (`PHP_OPCACHE_ENABLE=1`): the image reports it disabled by default and recommends enabling it for production |
| Caches | the image's autorun script runs `php artisan optimize` (config, routes, events, views) |
| Environment | `APP_ENV=production`, `APP_DEBUG=false`; `APP_KEY` is a throwaway benchmark value; logs go to standard error |
| API | controllers in `routes/api.php` under the `/api` prefix; `validator()` answers `422`, reshaped to the contract's `{"error": ...}` |
| API auth | Sanctum personal access tokens (`createToken`), checked against the database on every request, as Sanctum does by default |
| Staff account | `bench-staff`, a row of the default `users` table created by an entrypoint script with `tinker` |
| Admin | none (see `app.json`) |
| Container limits | 2 CPUs, 2 GiB, the same as every other app |
