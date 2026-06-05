# SonarQube Setup

SonarQube Community Edition is running for MergeMarket at:

`http://95.111.228.35:9000`

The tracked K3s manifest is `infra/k3s/sonarqube.yml`. It defines a single
SonarQube deployment plus persistent volumes for:

- `/opt/sonarqube/data`
- `/opt/sonarqube/extensions`
- `/opt/sonarqube/logs`

The service exposes SonarQube on NodePort `30900`, which maps to SonarQube's
container port `9000`. The VPS firewall should allow TCP `9000` for the public
instance currently in use. If the K3s NodePort path is used directly, allow TCP
`30900` or place NGINX in front and forward public `9000` to the service.

## Deploy or Reconcile

```bash
kubectl apply -f infra/k3s/sonarqube.yml
kubectl -n mergemarket-observability rollout status deploy/sonarqube
kubectl -n mergemarket-observability get pvc
```

## Persistence Check

Verify all three PVCs are `Bound`:

```bash
kubectl -n mergemarket-observability get pvc sonarqube-data sonarqube-extensions sonarqube-logs
```

Then restart the pod and confirm the instance keeps its projects, users, and
quality gate history:

```bash
kubectl -n mergemarket-observability rollout restart deploy/sonarqube
kubectl -n mergemarket-observability rollout status deploy/sonarqube
```

## Firewall

On the VPS:

```bash
sudo ufw allow 9000/tcp
sudo ufw status
```

## Generate the CI Token

1. Sign in to SonarQube as an administrator.
2. Open `My Account` -> `Security`.
3. Generate a user token named `jenkins-mergemarket`.
4. Store it in Jenkins credentials as `SONAR_TOKEN`.
5. Keep `.env` and Jenkins credentials private; never commit real tokens.

The example variable in `.env.example` is only a placeholder for local
development.
