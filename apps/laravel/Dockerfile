FROM composer:2 AS vendor
WORKDIR /app
COPY composer.json composer.lock ./
RUN composer install --no-dev --no-scripts --no-autoloader --prefer-dist --no-progress

FROM serversideup/php:8.4-fpm-nginx
ENV APP_ENV=production APP_DEBUG=false PHP_OPCACHE_ENABLE=1 AUTORUN_ENABLED=true SSL_MODE=off
WORKDIR /var/www/html
COPY --chown=www-data:www-data . .
COPY --from=vendor --chown=www-data:www-data /app/vendor ./vendor
USER root
COPY --from=vendor /usr/bin/composer /usr/bin/composer
RUN composer dump-autoload --optimize --no-dev --classmap-authoritative \
    && rm /usr/bin/composer \
    && install -m 0755 docker-staff.sh /etc/entrypoint.d/60-staff.sh \
    && chown -R www-data:www-data storage bootstrap/cache
USER www-data
EXPOSE 8080
