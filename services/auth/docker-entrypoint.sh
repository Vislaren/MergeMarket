#!/bin/sh
set -eu

: "${AUTH_TLS_CERT_FILE:=/certs/auth.local.crt}"
: "${AUTH_TLS_KEY_FILE:=/certs/auth.local.key}"
export AUTH_TLS_CERT_FILE AUTH_TLS_KEY_FILE

if [ ! -s "$AUTH_TLS_CERT_FILE" ] || [ ! -s "$AUTH_TLS_KEY_FILE" ]; then
  mkdir -p "$(dirname "$AUTH_TLS_CERT_FILE")" "$(dirname "$AUTH_TLS_KEY_FILE")"
  openssl req -x509 -newkey rsa:2048 -sha256 -days 365 -nodes \
    -keyout "$AUTH_TLS_KEY_FILE" \
    -out "$AUTH_TLS_CERT_FILE" \
    -subj "/CN=auth" \
    -addext "subjectAltName=DNS:auth,DNS:auth-service,DNS:localhost,IP:127.0.0.1" >/dev/null 2>&1
fi

exec /usr/local/bin/auth-service
