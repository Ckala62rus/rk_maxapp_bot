#!/bin/sh
set -eu

# Paths are expected to point to mounted certificate files.
#
# Defaults are for Let's Encrypt layout:
#   /etc/letsencrypt/live/<domain>/fullchain.pem
#   /etc/letsencrypt/live/<domain>/privkey.pem
CERT_DOMAIN="${CERT_DOMAIN:-}"
SSL_CERT_PATH_DEFAULT="/etc/letsencrypt/live/${CERT_DOMAIN}/fullchain.pem"
SSL_KEY_PATH_DEFAULT="/etc/letsencrypt/live/${CERT_DOMAIN}/privkey.pem"

SSL_CERT_PATH="${SSL_CERT_PATH:-$SSL_CERT_PATH_DEFAULT}"
SSL_KEY_PATH="${SSL_KEY_PATH:-$SSL_KEY_PATH_DEFAULT}"

# If cert files exist, enable SSL nginx config; otherwise keep HTTP-only config.
if [ -f "$SSL_CERT_PATH" ] && [ -f "$SSL_KEY_PATH" ]; then
  envsubst '${SSL_CERT_PATH} ${SSL_KEY_PATH}' \
    < /etc/nginx/templates/nginx.ssl.template.conf \
    > /etc/nginx/conf.d/default.conf
fi

exec nginx -g 'daemon off;'

