# K3s Deployment

The VPS can run MergeMarket two ways:

- Docker Compose: `docker compose build && docker compose up -d`
- K3s: `sh infra/scripts/deploy-k3s.sh`

The K3s script builds the implemented Go service images locally, imports them
into the K3s containerd image store, creates required ConfigMaps/Secrets, then
applies `infra/k3s/mergemarket-apps.yml`.

## Before Running

Create `.env` from `.env.example` and set real secrets:

```bash
cp .env.example .env
```

Required values:

- `DB_PASSWORD`
- `JWT_SECRET`
- `AUTH_ENCRYPTION_KEY`
- `REDIS_PASSWORD` may be empty, but should be set on the VPS

## Deploy

```bash
sh infra/scripts/deploy-k3s.sh
kubectl -n mergemarket get pods
kubectl -n mergemarket get svc
```

Kong is exposed through NodePort `30088` for proxy traffic and `30081` for the
admin API. Put NGINX/Coolify in front of `30088` if the public port should remain
`8088`.

## Notes

- `services/bff/` is not deployed yet because it currently only contains a
  README. Kong routes directly to `search`, `userdata`, and `history` until the
  BFF service is implemented.
- `GET /api/v1/products/{id}/truth-score` is still waiting for a real backend
  implementation; the mock server has a fixture endpoint for client development.
- SonarQube and Grafana keep their own manifests in `infra/k3s/sonarqube.yml`
  and `infra/k3s/grafana.yml`.
