#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
cd "$ROOT_DIR"

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

: "${DB_PASSWORD:?DB_PASSWORD must be set in .env or the environment}"
: "${REDIS_PASSWORD:=}"
: "${JWT_SECRET:?JWT_SECRET must be set in .env or the environment}"
: "${AUTH_ENCRYPTION_KEY:?AUTH_ENCRYPTION_KEY must be set in .env or the environment}"

kubectl create namespace mergemarket --dry-run=client -o yaml | kubectl apply -f -

kubectl -n mergemarket create secret generic mergemarket-secrets \
  --from-literal=DB_PASSWORD="$DB_PASSWORD" \
  --from-literal=REDIS_PASSWORD="$REDIS_PASSWORD" \
  --from-literal=JWT_SECRET="$JWT_SECRET" \
  --from-literal=AUTH_ENCRYPTION_KEY="$AUTH_ENCRYPTION_KEY" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n mergemarket create configmap postgres-initdb \
  --from-file=01-schema.sql=infra/db/init/01-schema.sql \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n mergemarket create configmap scraper-configs \
  --from-file=services/scraper-service/configs \
  --dry-run=client -o yaml | kubectl apply -f -

secret_escaped="$(printf '%s' "$JWT_SECRET" | sed 's/[\/&]/\\&/g')"
tmp_kong="$(mktemp)"
sed "s|__JWT_SECRET__|$secret_escaped|g" api-gateway/kong.yml > "$tmp_kong"
kubectl -n mergemarket create configmap kong-config \
  --from-file=kong.yml="$tmp_kong" \
  --dry-run=client -o yaml | kubectl apply -f -
rm -f "$tmp_kong"

docker build -t mergemarket/auth:local services/auth
docker build -t mergemarket/proxy-validator:local services/proxy-validator
docker build -t mergemarket/scraper:local services/scraper-service
docker build -t mergemarket/normalization:local services/normalization
docker build -t mergemarket/history:local services/history
docker build -t mergemarket/search:local services/search
docker build -t mergemarket/userdata:local services/userdata

docker save \
  mergemarket/auth:local \
  mergemarket/proxy-validator:local \
  mergemarket/scraper:local \
  mergemarket/normalization:local \
  mergemarket/history:local \
  mergemarket/search:local \
  mergemarket/userdata:local \
  | sudo k3s ctr images import -

kubectl apply -f infra/k3s/mergemarket-apps.yml
kubectl -n mergemarket rollout status deploy/auth
kubectl -n mergemarket rollout status deploy/search
kubectl -n mergemarket rollout status deploy/userdata
kubectl -n mergemarket rollout status deploy/kong
