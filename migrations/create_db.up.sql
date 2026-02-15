CREATE USER app_url
    PASSWORD 'password';

CREATE DATABASE url_shortener
    OWNER 'app'
    ENCODING 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8';