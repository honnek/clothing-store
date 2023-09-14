#!/usr/bin/env bash
export APP_ENV=test

php bin/console doctrine:database:drop --if-exists --force || true

php bin/console doctrine:database:create

php bin/console doctrine:schema:update --dump-sql
php bin/console doctrine:schema:update --force
php bin/console hautelook:fixtures:load -n

php ./vendor/bin/phpunit --testdox