FROM php:8.1-fpm

WORKDIR /var/www

RUN apt update
RUN docker-php-ext-install exif
RUN apt-get install libmagickwand-dev libmagickcore-dev -y
RUN pecl install imagick
RUN docker-php-ext-enable imagick
RUN PHP_OPENSSL=yes
RUN docker-php-ext-install xml
RUN docker-php-ext-install filter
RUN apt-get install libzip-dev -y
RUN docker-php-ext-install zip
RUN docker-php-ext-install bcmath
RUN apt install libwebp-dev
RUN docker-php-ext-configure gd --with-freetype --with-jpeg --with-webp
RUN docker-php-ext-install gd
RUN docker-php-ext-install intl

CMD ["php-fpm"]

