#!/bin/sh
set -eu

CERT_DOMAIN="${CERT_DOMAIN:-}"

if [ -n "$CERT_DOMAIN" ] && [ -f "/etc/letsencrypt/live/${CERT_DOMAIN}/fullchain.pem" ]; then
    # SSL режим - подставляем переменные в шаблон
    export SSL_CERT_PATH="/etc/letsencrypt/live/${CERT_DOMAIN}/fullchain.pem"
    export SSL_KEY_PATH="/etc/letsencrypt/live/${CERT_DOMAIN}/privkey.pem"

    envsubst '${SSL_CERT_PATH} ${SSL_KEY_PATH} ${CERT_DOMAIN}' \
        < /etc/nginx/templates/nginx.ssl.template.conf \
        > /etc/nginx/conf.d/default.conf
else
    # HTTP режим - копируем обычный конфиг
    cp /etc/nginx/conf.d/default.conf.bak /etc/nginx/conf.d/default.conf 2>/dev/null || true
fi

exec nginx -g 'daemon off;'
