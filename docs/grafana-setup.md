# Grafana Setup (A-12)

Grafana provides the MergeMarket **quality & pipeline** dashboard, built from
the **SonarQube API** and the **GitHub Actions API** via the open-source
[Infinity datasource](https://grafana.com/grafana/plugins/yesoreyeram-infinity-datasource/)
(neither API has a native Grafana datasource). It runs on the VPS at:

`http://95.111.228.35:3000`

The tracked artefacts are:

| Path | Purpose |
|---|---|
| `infra/k3s/grafana.yml` | K3s deployment, PVC, provisioning ConfigMaps, NodePort service |
| `infra/grafana/provisioning/datasources/datasources.yml` | SonarQube + GitHub Infinity datasources |
| `infra/grafana/provisioning/dashboards/provider.yml` | File-based dashboard provider |
| `infra/grafana/dashboards/mergemarket-quality.json` | The dashboard (7 panels) |

Grafana is deployed into the shared `mergemarket-observability` namespace
(created by `infra/k3s/sonarqube.yml`, A-11) and exposed on NodePort `30300`
mapping to container port `3000`.

> The datasource provisioning is duplicated inside `grafana.yml` as ConfigMaps
> (`grafana-datasources`, `grafana-dashboard-provider`) so the manifest is
> self-contained. If you edit the files under `infra/grafana/provisioning/`,
> mirror the change into `grafana.yml`.

## Dashboard panels

| Panel | Source | Metric |
|---|---|---|
| Test Coverage % | SonarQube `/api/measures/component` | `coverage` |
| Open Bugs | SonarQube `/api/measures/component` | `bugs` |
| Vulnerabilities | SonarQube `/api/measures/component` | `vulnerabilities` |
| Pipeline Pass/Fail Rate | GitHub `/repos/{repo}/actions/runs` | `conclusion` counts |
| Coverage % Over Time | SonarQube `/api/measures/search_history` | `coverage` |
| Bugs & Vulnerabilities Over Time | SonarQube `/api/measures/search_history` | `bugs`, `vulnerabilities` |
| Pipeline Build Duration Over Time | GitHub `/repos/{repo}/actions/runs` | `updated_at − run_started_at` |

Two dashboard variables let you retarget without editing JSON:
`sonar_project` (default `mergemarket`) and `gh_repo`
(default `Vislaren/MergeMarket`).

## 1. Create the secret (do this first — not committed)

Grafana reads its admin credentials and the API tokens from a Kubernetes
Secret named `grafana-secrets`. Create it on the VPS with real values:

```bash
kubectl -n mergemarket-observability create secret generic grafana-secrets \
  --from-literal=admin-user='admin' \
  --from-literal=admin-password='<strong-password>' \
  --from-literal=sonar-token='<sonarqube-user-token>' \
  --from-literal=github-token='<github-fine-grained-PAT>'
```

- **SonarQube token** — generate in SonarQube under *My Account → Security*
  (a read token is sufficient). The datasource sends it as the HTTP Basic
  *username* with an empty password.
- **GitHub PAT** — a fine-grained token with **read-only** access to
  *Actions* and repository *metadata* on `Vislaren/MergeMarket`. Sent as a
  Bearer token.

To rotate later: `kubectl -n mergemarket-observability delete secret
grafana-secrets` then recreate, and `kubectl -n mergemarket-observability
rollout restart deploy/grafana`.

## 2. Create the dashboard ConfigMap

The dashboard JSON is loaded from a ConfigMap generated from the tracked file
(kept out of the manifest because it is large):

```bash
kubectl -n mergemarket-observability create configmap grafana-dashboards \
  --from-file=infra/grafana/dashboards/ \
  --dry-run=client -o yaml | kubectl apply -f -
```

Re-run this exact command whenever the dashboard JSON changes, then restart the
pod (step 4).

## 3. Deploy

```bash
kubectl apply -f infra/k3s/grafana.yml
kubectl -n mergemarket-observability rollout status deploy/grafana
kubectl -n mergemarket-observability get pvc grafana-data
```

## 4. Reconcile / apply updates

```bash
kubectl -n mergemarket-observability rollout restart deploy/grafana
kubectl -n mergemarket-observability rollout status deploy/grafana
```

## 5. Firewall

```bash
sudo ufw allow 30300/tcp     # direct NodePort access
sudo ufw status
```

To serve Grafana on the public port `3000` instead, place NGINX in front and
forward `3000 → 30300` (matching the SonarQube `9000` pattern in
`docs/sonarqube-setup.md`).

## Local Docker alternative (optional)

For a quick local look without K3s:

```bash
docker run -d --name grafana -p 3000:3000 \
  -e GF_INSTALL_PLUGINS=yesoreyeram-infinity-datasource \
  -e GF_SECURITY_ADMIN_PASSWORD="$GRAFANA_ADMIN_PASSWORD" \
  -e SONAR_HOST_URL="$SONAR_HOST_URL" \
  -e SONAR_TOKEN="$SONAR_TOKEN" \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -v "$PWD/infra/grafana/provisioning:/etc/grafana/provisioning" \
  -v "$PWD/infra/grafana/dashboards:/var/lib/grafana/dashboards" \
  grafana/grafana:11.1.0
```

Set the variables in your `.env` first (see the Grafana block in
`.env.example`).

## Known limitations / follow-ups

- **Pipeline panels depend on GitHub Actions data.** The "Pipeline Pass/Fail
  Rate" and "Build Duration" panels query the GitHub Actions *workflow runs*
  API. CI was migrated to **Jenkins** (`Jenkinsfile`, A-10) and the GitHub
  Actions workflow directory was deleted in a prior session, so these two
  panels will be **empty until an Actions workflow runs again**. Options:
  (a) restore a GitHub Actions workflow alongside Jenkins, or (b) re-point
  these panels at a Jenkins datasource (the
  [Jenkins data source / Infinity over the Jenkins JSON API]). Tracked as a
  CI-source reconciliation item also noted in the A-04/A-10 DONE entries.
- **SonarQube project key.** The dashboard defaults `sonar_project=mergemarket`.
  Confirm this matches `sonar.projectKey` once `sonar-project.properties` is
  finalised (A-10/A-11), and adjust the variable default if it differs.
- **The live VPS was not modified from this local session** (the VPS is
  unreachable from local dev). The manifest, provisioning, dashboard, and this
  runbook are committed for application on the VPS — same approach as the A-11
  SonarQube deliverable.
