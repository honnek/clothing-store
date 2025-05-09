FROM php:8.2-fpm-alpine

# Copy php config
COPY ./conf/php.ini /usr/local/etc/php/php.ini

RUN apk update

# Packages
RUN apk add --no-cache curl git build-base zlib-dev oniguruma-dev autoconf icu-dev bash php-bcmath php-sockets

RUN apk update && rm -rf /var/cache/apk/*

# Postgres
#RUN apk add --no-cache libpq-dev \
#    && docker-php-ext-install pdo_pgsql
RUN set -ex \
  && apk --no-cache add \
    postgresql-dev

RUN docker-php-ext-install pdo pdo_pgsql

# Intl
RUN docker-php-ext-configure intl \
    && docker-php-ext-install intl

RUN docker-php-ext-install bcmath
RUN apk add --no-cache linux-headers
RUN docker-php-ext-install sockets

# OPCache (if you need it by any means)
# RUN docker-php-ext-install opcache

# XDebug
ARG INSTALL_XDEBUG=true

RUN if [ ${INSTALL_XDEBUG} = true ]; \
  then \
    apk add --no-cache linux-headers  \
    && apk add --update --no-cache --virtual .build-dependencies $PHPIZE_DEPS \
    && pecl install xdebug \
    && docker-php-ext-enable xdebug \
    && pecl clear-cache \
    && apk del .build-dependencies; \
fi;


# Xdebug config
COPY ./conf/xdebug.ini /usr/local/etc/php/conf.d/xdebug.ini

# Non root user
ARG PUID=1000
ARG PGID=1000
RUN apk --no-cache add shadow && \
    groupmod -o -g ${PGID} www-data && \
    usermod -o -u ${PUID} -g www-data www-data

# Composer
RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer

# Symfony CLI
RUN curl -sS https://get.symfony.com/cli/installer | bash
RUN mv /root/.symfony5/bin/symfony /usr/local/bin/symfony

# Git config
RUN git config --global user.email "email@domain.com" \
    && git config --global user.name "Your Name"

# Setup GD extension
RUN apk add --no-cache \
      freetype \
      libjpeg-turbo \
      libpng \
      freetype-dev \
      libjpeg-turbo-dev \
      libpng-dev \
    && docker-php-ext-configure gd \
      --with-freetype=/usr/include/ \
      # --with-png=/usr/include/ \ # No longer necessary as of 7.4; https://github.com/docker-library/php/pull/910#issuecomment-559383597
      --with-jpeg=/usr/include/ \
    && docker-php-ext-install -j$(nproc) gd \
    && docker-php-ext-enable gd \
    && apk del --no-cache \
      freetype-dev \
      libjpeg-turbo-dev \
      libpng-dev \
    && rm -rf /tmp/*

# Setup ext-zip php extension
RUN apk add --no-cache \
      libzip-dev \
      zip \
    && docker-php-ext-install zip

# Setup chromium and panther
#RUN apk add --no-cache \
#        chromium \
#        chromium-chromedriver
ENV PANTHER_CHROME_DRIVER_BINARY /usr/lib/chromium/chromedriver
ENV PANTHER_NO_SANDBOX 1
ENV PANTHER_WEB_SERVER_PORT 9800


# Set user
USER www-data

# Files
COPY ./ /var/www
WORKDIR /var/www

# Run php-fpm
CMD php-fpm

# Expose port
EXPOSE 9000